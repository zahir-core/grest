package grest

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// Base attributes
const (
	FmtReset uint8 = iota
	FmtBold
	FmtFaint
	FmtItalic
	FmtUnderline
	FmtBlinkSlow
	FmtBlinkRapid
	FmtReverseVideo
	FmtConcealed
	FmtCrossedOut
)

// Foreground text colors
const (
	FmtBlack uint8 = iota + 30
	FmtRed
	FmtGreen
	FmtYellow
	FmtBlue
	FmtMagenta
	FmtCyan
	FmtWhite
)

// Foreground Hi-Intensity text colors
const (
	FmtHiBlack uint8 = iota + 90
	FmtHiRed
	FmtHiGreen
	FmtHiYellow
	FmtHiBlue
	FmtHiMagenta
	FmtHiCyan
	FmtHiWhite
)

// Background text colors
const (
	FmtBgBlack uint8 = iota + 40
	FmtBgRed
	FmtBgGreen
	FmtBgYellow
	FmtBgBlue
	FmtBgMagenta
	FmtBgCyan
	FmtBgWhite
)

// Background Hi-Intensity text colors
const (
	FmtBgHiBlack uint8 = iota + 100
	FmtBgHiRed
	FmtBgHiGreen
	FmtBgHiYellow
	FmtBgHiBlue
	FmtBgHiMagenta
	FmtBgHiCyan
	FmtBgHiWhite
)

var (
	// DisableFmt defines if the output is colorized or not.
	DisableFmt = (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))

	// Output defines the standard output of the print functions. By default os.Stdout is used.
	FmtStdout = colorable.NewColorableStdout()
)

// Fmt format log with attribute
//	for example :
//	  log.Fmt("text", log.Bold, log.Red)
//
//	output (text with bold red foreground) :
//	  \x1b[1;31mtext\x1b[0m
func Fmt(s string, attribute ...uint8) string {
	if DisableFmt {
		return s
	}
	format := make([]string, len(attribute))
	for i, v := range attribute {
		format[i] = strconv.Itoa(int(v))
	}
	return "\x1b[" + strings.Join(format, ";") + "m" + s + "\x1b[0m"
}

type FileFormatter struct{}

type StructTag struct {
	Key   string
	Value string
}

func FormatFile(paths ...string) {
	if len(paths) == 0 {
		paths = append(paths, ".")
	}
	f := FileFormatter{}
	for _, path := range paths {
		filepath.Walk(path,
			func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(p, ".go") {
					f.Format(p)
				}
				return nil
			})
	}
}

func (ff FileFormatter) Format(fileName string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range f.Decls {
		genDecl, isGenDecl := node.(*ast.GenDecl)
		if isGenDecl {
			for _, spec := range genDecl.Specs {
				typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
				if isTypeSpec {
					fmt.Println("Formatting", fileName, typeSpec.Name.Name)
					structType, isStructType := typeSpec.Type.(*ast.StructType)
					if isStructType {
						mapTag, maxTagLen := ff.ParseTag(structType.Fields.List)
						ff.RewriteTag(structType.Fields.List, mapTag, maxTagLen)
					}
				}
			}
		}
	}
	var buf bytes.Buffer
	err = format.Node(&buf, fset, f)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fileName, buf.Bytes(), 0)
	if err != nil {
		log.Fatal(err)
	}
}

func (ff FileFormatter) ParseTag(fields []*ast.Field) (mapTag map[string][]StructTag, maxTagLen map[string]int) {
	mapTag = map[string][]StructTag{}
	maxTagLen = map[string]int{}
	for _, field := range fields {
		if len(field.Names) > 0 {
			if field.Tag == nil {
				continue
			}

			var tags []StructTag
			ftv, _ := strconv.Unquote(field.Tag.Value)
			tgs := strings.Split(strings.ReplaceAll(ftv, `:"`, "==="), `"`)
			for _, tg := range tgs {
				t := strings.Split(strings.Trim(tg, " "), "===")
				if len(t) > 1 {
					key := t[0]
					value := t[1]
					lenVal := len(value)
					ml, isMaxLenExist := maxTagLen[key]
					if !isMaxLenExist || lenVal > ml {
						maxTagLen[key] = lenVal
					}
					tags = append(tags, StructTag{Key: key, Value: value})
				}
			}
			mapTag[field.Names[0].Name] = tags
		}
	}

	return mapTag, maxTagLen
}

func (ff FileFormatter) RewriteTag(fields []*ast.Field, mapTag map[string][]StructTag, maxTagLen map[string]int) {
	for _, field := range fields {
		if len(field.Names) > 0 {
			tags, isExist := mapTag[field.Names[0].Name]
			if isExist {
				if field.Tag == nil {
					field.Tag = &ast.BasicLit{}
				}
				field.Tag.Value = ff.FormattedTagString(tags, maxTagLen)
			}
		}
	}
}

func (ff FileFormatter) FormattedTagString(tags []StructTag, maxTagLen map[string]int) string {
	if len(tags) == 0 {
		return ""
	}
	sortedTags := []StructTag{}
	sort.Slice(tags, func(i, j int) bool { return tags[i].Key < tags[j].Key })
	for _, tagKey := range []string{"json", "form", "xml", "db", "gorm", "validate", "default", "example", "title", "note"} {
		for _, tag := range tags {
			if tag.Key == tagKey {
				sortedTags = append(sortedTags, tag)
			}
		}
	}

	for _, tag := range tags {
		switch tag.Key {
		case "json", "form", "xml", "db", "gorm", "validate", "default", "example", "title", "note":
			// do nothing
		default:
			// append additional tag
			sortedTags = append(sortedTags, tag)
		}
	}
	newTag := ""
	for _, t := range sortedTags {
		newTag += t.Key + ":" + ff.TagValueWithDelimiter(t.Value, maxTagLen[t.Key])
	}
	return "`" + strings.Trim(newTag, " ") + "`"
}

func (ff FileFormatter) TagValueWithDelimiter(str string, maxLen int) string {
	tag := `"` + str + `"`
	n := maxLen - len(str) + 1
	for i := 0; i < n; i++ {
		tag += " "
	}
	return tag
}
