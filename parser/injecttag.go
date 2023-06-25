package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/ml444/glog"
)

var (
	rComment = regexp.MustCompile(`^//.*?@([\w_]+?):\s*(.*)$`)
	rInject  = regexp.MustCompile("`.+`$")
	rTags    = regexp.MustCompile(`[\w_]+:"[^"]+"`)
	rAll     = regexp.MustCompile(".*")
)

type textArea struct {
	Start        int
	End          int
	CurrentTag   string
	InjectTag    string
	CommentStart int
	CommentEnd   int
}

type tagItem struct {
	key   string
	value string
}

type tagItems []tagItem

func (ti tagItems) format() string {
	var tags []string
	for _, item := range ti {
		tags = append(tags, fmt.Sprintf(`%s:%s`, item.key, item.value))
	}
	return strings.Join(tags, " ")
}

func injectTag(contents []byte, area textArea, removeTagComment bool) (injected []byte) {
	newTagItems := func(tag string) tagItems {
		var items []tagItem
		splitList := rTags.FindAllString(tag, -1)

		for _, t := range splitList {
			sepPos := strings.Index(t, ":")
			items = append(items, tagItem{
				key:   t[:sepPos],
				value: t[sepPos+1:],
			})
		}
		return items
	}

	merge := func(ti, nti tagItems) tagItems {
		var results []tagItem
		for i := range ti {
			dup := -1
			for j := range nti {
				if ti[i].key == nti[j].key {
					dup = j
					break
				}
			}
			if dup == -1 {
				results = append(results, ti[i])
			} else {
				results = append(results, nti[dup])
				nti = append(nti[:dup], nti[dup+1:]...)
			}
		}
		return append(results, nti...)
	}

	expr := make([]byte, area.End-area.Start)
	copy(expr, contents[area.Start-1:area.End-1])
	cti := newTagItems(area.CurrentTag)
	iti := newTagItems(area.InjectTag)
	ti := merge(cti, iti)
	expr = rInject.ReplaceAll(expr, []byte(fmt.Sprintf("`%s`", ti.format())))

	if removeTagComment {
		strippedComment := make([]byte, area.CommentEnd-area.CommentStart)
		copy(strippedComment, contents[area.CommentStart-1:area.CommentEnd-1])
		strippedComment = rAll.ReplaceAll(expr, []byte(" "))
		if area.CommentStart < area.Start {
			injected = append(injected, contents[:area.CommentStart-1]...)
			injected = append(injected, strippedComment...)
			injected = append(injected, contents[area.CommentEnd-1:area.Start-1]...)
			injected = append(injected, expr...)
			injected = append(injected, contents[area.End-1:]...)
		} else {
			injected = append(injected, contents[:area.Start-1]...)
			injected = append(injected, expr...)
			injected = append(injected, contents[area.End-1:area.CommentStart-1]...)
			injected = append(injected, strippedComment...)
			injected = append(injected, contents[area.CommentEnd-1:]...)
		}
	} else {
		injected = append(injected, contents[:area.Start-1]...)
		injected = append(injected, expr...)
		injected = append(injected, contents[area.End-1:]...)
	}

	return
}

func ParsePbFile(inputPath string, src interface{}, xxxSkip []string) (areas []textArea, err error) {
	log.Infof("parsing file %q for inject tag comments", inputPath)
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, inputPath, src, parser.ParseComments)
	if err != nil {
		return
	}

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		var typeSpec *ast.TypeSpec
		for _, spec := range genDecl.Specs {
			if ts, tsOK := spec.(*ast.TypeSpec); tsOK {
				typeSpec = ts
				break
			}
		}

		// skip if can't get type spec
		if typeSpec == nil {
			continue
		}

		// not a struct, skip
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		builder := strings.Builder{}
		if len(xxxSkip) > 0 {
			for i, skip := range xxxSkip {
				builder.WriteString(fmt.Sprintf("%s:\"-\"", skip))
				if i > 0 {
					builder.WriteString(",")
				}
			}
		}

		for _, field := range structDecl.Fields.List {
			// skip if field has no doc
			if len(field.Names) > 0 {
				name := field.Names[0].Name
				if len(xxxSkip) > 0 && strings.HasPrefix(name, "XXX") {
					currentTag := field.Tag.Value
					area := textArea{
						Start:      int(field.Pos()),
						End:        int(field.End()),
						CurrentTag: currentTag[1 : len(currentTag)-1],
						InjectTag:  builder.String(),
					}
					areas = append(areas, area)
				}
			}

			var comments []*ast.Comment

			if field.Doc != nil {
				comments = append(comments, field.Doc.List...)
			}

			if field.Comment != nil {
				comments = append(comments, field.Comment.List...)
			}

			for _, comment := range comments {
				match := rComment.FindStringSubmatch(comment.Text)
				if len(match) != 3 {
					continue
				}
				tag := fmt.Sprintf(`%s:"%s"`, match[1], match[2])

				currentTag := field.Tag.Value
				area := textArea{
					Start:        int(field.Pos()),
					End:          int(field.End()),
					CurrentTag:   currentTag[1 : len(currentTag)-1],
					InjectTag:    tag,
					CommentStart: int(comment.Pos()),
					CommentEnd:   int(comment.End()),
				}
				areas = append(areas, area)
			}
		}
	}
	log.Infof("parsed file %q, number of fields to inject custom tags: %d", inputPath, len(areas))
	return
}

func WritePbFile(inputPath string, areas []textArea, removeTagComment bool) (err error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return
	}

	contents, err := io.ReadAll(f)
	if err != nil {
		return
	}

	if err = f.Close(); err != nil {
		return
	}

	// inject custom tags from tail of file first to preserve order
	for i := range areas {
		area := areas[len(areas)-i-1]
		println(fmt.Sprintf("==> inject custom tag %q ", area.InjectTag))
		contents = injectTag(contents, area, removeTagComment)
	}
	if err = os.WriteFile(inputPath, contents, 0o644); err != nil {
		return
	}

	if len(areas) > 0 {
		log.Infof("file %q is injected with custom tags", inputPath)
	}
	return
}
