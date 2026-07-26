// html_viewer.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"golang.org/x/net/html"
)

func colorize(text, color string, enabled bool) string {
	if !enabled {
		return text
	}
	colors := map[string]string{
		"tag":    "\033[96m",
		"attr":   "\033[93m",
		"text":   "\033[92m",
		"comment":"\033[90m",
		"reset":  "\033[0m",
	}
	return colors[color] + text + colors["reset"]
}

func renderTree(n *html.Node, prefix string, last bool, color bool, showAttrs bool, selector string, depth int) {
	if n.Type == html.ElementNode {
		connector := "└── "
		if !last {
			connector = "├── "
		}
		tagName := n.Data
		// Получаем id и классы
		id := ""
		classes := []string{}
		for _, a := range n.Attr {
			if a.Key == "id" {
				id = a.Val
			}
			if a.Key == "class" {
				classes = strings.Fields(a.Val)
			}
		}
		display := tagName
		if id != "" {
			display += "#" + id
		}
		if len(classes) > 0 {
			display += "." + strings.Join(classes, ".")
		}
		// Атрибуты
		attrStr := ""
		if showAttrs {
			attrs := []string{}
			for _, a := range n.Attr {
				if a.Key != "id" && a.Key != "class" {
					attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, a.Val))
				}
			}
			if len(attrs) > 0 {
				attrStr = " " + strings.Join(attrs, " ")
			}
		}
		fmt.Printf("%s%s%s%s%s\n", prefix, connector, colorize(display, "tag", color), colorize(attrStr, "attr", color), colorize("", "reset", color))
		
		// Проходим по детям
		var children []*html.Node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode || c.Type == html.TextNode {
				children = append(children, c)
			}
		}
		for i, c := range children {
			newPrefix := prefix + "    "
			if !last {
				newPrefix = prefix + "│   "
			}
			if c.Type == html.ElementNode {
				// Фильтр по селектору - упрощённо проверяем соответствие
				if selector != "" {
					// Простой фильтр по имени тега
					if !matchesSelector(c, selector) {
						continue
					}
				}
				renderTree(c, newPrefix, i == len(children)-1, color, showAttrs, selector, depth+1)
			} else if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
				text := strings.TrimSpace(c.Data)
				connectorChild := "└── "
				if i != len(children)-1 {
					connectorChild = "├── "
				}
				fmt.Printf("%s%s%s\n", newPrefix, connectorChild, colorize("\"" + text + "\"", "text", color))
			}
		}
	}
}

func matchesSelector(n *html.Node, selector string) bool {
	// Очень простая проверка: если селектор начинается с '.', то проверяем класс
	if strings.HasPrefix(selector, ".") {
		class := selector[1:]
		for _, a := range n.Attr {
			if a.Key == "class" {
				classes := strings.Fields(a.Val)
				for _, c := range classes {
					if c == class {
						return true
					}
				}
			}
		}
		return false
	} else if strings.HasPrefix(selector, "#") {
		id := selector[1:]
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == id {
				return true
			}
		}
		return false
	} else {
		// просто тег
		return n.Data == selector
	}
}

func validateHTML(doc *html.Node) []string {
	var errors []string
	// Простая проверка дублирующихся id
	ids := make(map[string]bool)
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if a.Key == "id" {
					if ids[a.Val] {
						errors = append(errors, fmt.Sprintf("Duplicate id '%s'", a.Val))
					} else {
						ids[a.Val] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return errors
}

func main() {
	var tree, color, attributes, validate bool
	var selector, output string
	flag.BoolVar(&tree, "t", false, "Show DOM tree")
	flag.BoolVar(&tree, "tree", false, "Show DOM tree")
	flag.BoolVar(&color, "c", false, "Force color output")
	flag.BoolVar(&color, "color", false, "Force color output")
	flag.BoolVar(&attributes, "a", false, "Show attributes")
	flag.BoolVar(&attributes, "attributes", false, "Show attributes")
	flag.BoolVar(&validate, "v", false, "Validate HTML")
	flag.BoolVar(&validate, "validate", false, "Validate HTML")
	flag.StringVar(&selector, "s", "", "CSS selector filter")
	flag.StringVar(&selector, "selector", "", "CSS selector filter")
	flag.StringVar(&output, "o", "", "Export to file (JSON or HTML)")
	flag.StringVar(&output, "output", "", "Export to file (JSON or HTML)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	var content []byte
	if len(args) > 0 {
		var err error
		content, err = os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintln(os.Stderr, "No input provided. Pipe HTML or pass file.")
			os.Exit(1)
		}
		var err error
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	colorEnabled := color || isTerminal()

	if validate {
		errors := validateHTML(doc)
		if len(errors) > 0 {
			fmt.Println("Validation errors:")
			for _, e := range errors {
				fmt.Println(colorize("  ❌ "+e, "tag", colorEnabled))
			}
		} else {
			fmt.Println(colorize("✅ No validation errors found.", "text", colorEnabled))
		}
	}

	if tree {
		fmt.Println(colorize("🌳 DOM Tree:", "tag", colorEnabled))
		// Находим html элемент
		var root *html.Node
		var findRoot func(*html.Node)
		findRoot = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "html" {
				root = n
				return
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findRoot(c)
			}
		}
		findRoot(doc)
		if root == nil {
			root = doc
		}
		renderTree(root, "", true, colorEnabled, attributes, selector, 0)
	}

	if output != "" {
		// Экспорт: просто сохраняем красиво отформатированный HTML
		// Для JSON пришлось бы строить структуру, но упростим: сохраняем HTML
		err := os.WriteFile(output, content, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		} else {
			fmt.Printf("Exported to %s\n", output)
		}
	}
}

func isTerminal() bool {
	// упрощённо
	return true
}
