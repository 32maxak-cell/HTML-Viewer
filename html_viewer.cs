// html_viewer.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using HtmlAgilityPack;

class HtmlViewer
{
    private static bool colorEnabled;

    static void Main(string[] args)
    {
        bool tree = false;
        bool color = false;
        bool attributes = false;
        bool validate = false;
        string selector = null;
        string output = null;
        string filePath = null;

        for (int i = 0; i < args.Length; i++)
        {
            switch (args[i])
            {
                case "-t":
                case "--tree":
                    tree = true;
                    break;
                case "-c":
                case "--color":
                    color = true;
                    break;
                case "-a":
                case "--attributes":
                    attributes = true;
                    break;
                case "-v":
                case "--validate":
                    validate = true;
                    break;
                case "-s":
                case "--selector":
                    if (i + 1 < args.Length) selector = args[++i];
                    else { Console.Error.WriteLine("Missing selector"); Environment.Exit(1); }
                    break;
                case "-o":
                case "--output":
                    if (i + 1 < args.Length) output = args[++i];
                    else { Console.Error.WriteLine("Missing output file"); Environment.Exit(1); }
                    break;
                case "-h":
                case "--help":
                    Console.WriteLine($"Usage: dotnet run -- [options] [file]");
                    Console.WriteLine("  -t, --tree          Show DOM tree");
                    Console.WriteLine("  -c, --color         Force color output");
                    Console.WriteLine("  -a, --attributes    Show attributes");
                    Console.WriteLine("  -v, --validate      Validate HTML");
                    Console.WriteLine("  -s, --selector      CSS selector filter");
                    Console.WriteLine("  -o, --output        Export to file");
                    Environment.Exit(0);
                    break;
                default:
                    if (filePath == null) filePath = args[i];
                    else { Console.Error.WriteLine($"Extra argument: {args[i]}"); Environment.Exit(1); }
                    break;
            }
        }

        colorEnabled = color || !Console.IsOutputRedirected;

        string content;
        if (filePath != null)
        {
            try { content = File.ReadAllText(filePath); }
            catch (Exception ex) { Console.Error.WriteLine($"Error reading file: {ex.Message}"); Environment.Exit(1); return; }
        }
        else
        {
            if (Console.IsInputRedirected)
            {
                using var reader = new StreamReader(Console.OpenStandardInput());
                content = reader.ReadToEnd();
            }
            else
            {
                Console.Error.WriteLine("No input provided. Pipe HTML or pass file.");
                Environment.Exit(1);
                return;
            }
        }

        var doc = new HtmlDocument();
        doc.LoadHtml(content);

        if (validate)
        {
            var errors = ValidateHtml(doc);
            if (errors.Count == 0)
                Console.WriteLine(Colorize("✅ No validation errors found.", "text"));
            else
            {
                Console.WriteLine("Validation errors:");
                foreach (var err in errors)
                    Console.WriteLine(Colorize($"  ❌ {err}", "tag"));
            }
        }

        if (tree)
        {
            Console.WriteLine(Colorize("🌳 DOM Tree:", "tag"));
            var root = doc.DocumentNode;
            // Находим html
            var htmlNode = root.SelectSingleNode("html");
            if (htmlNode != null)
                RenderTree(htmlNode, "", true, attributes, selector);
            else
                RenderTree(root, "", true, attributes, selector);
        }

        if (output != null)
        {
            // Просто сохраняем отформатированный HTML
            using var writer = new StreamWriter(output);
            doc.Save(writer);
            Console.WriteLine($"Exported to {output}");
        }
    }

    static void RenderTree(HtmlNode node, string prefix, bool last, bool showAttrs, string selector)
    {
        if (node.NodeType == HtmlNodeType.Element)
        {
            string connector = last ? "└── " : "├── ";
            string tagName = node.Name;
            string id = node.GetAttributeValue("id", "");
            string classAttr = node.GetAttributeValue("class", "");
            string display = tagName;
            if (!string.IsNullOrEmpty(id))
                display += "#" + id;
            if (!string.IsNullOrEmpty(classAttr))
            {
                var classes = classAttr.Split(' ', StringSplitOptions.RemoveEmptyEntries);
                if (classes.Length > 0)
                    display += "." + string.Join(".", classes);
            }
            string attrStr = "";
            if (showAttrs)
            {
                var attrs = node.Attributes
                    .Where(a => a.Name != "id" && a.Name != "class")
                    .Select(a => $"{a.Name}={a.Value}");
                if (attrs.Any())
                    attrStr = " " + string.Join(" ", attrs);
            }

            Console.WriteLine($"{prefix}{connector}{Colorize(display, "tag")}{Colorize(attrStr, "attr")}");

            // Фильтр по селектору: проверяем, подходит ли элемент или его дети
            bool matchesFilter = string.IsNullOrEmpty(selector) || 
                node.SelectSingleNode(selector) != null ||
                node.SelectNodes(selector)?.Any() == true;
            if (!matchesFilter && !string.IsNullOrEmpty(selector))
                return; // пропускаем, но дети могут подходить?

            // Собираем дочерние элементы и текстовые узлы
            var children = node.ChildNodes.Where(n => n.NodeType == HtmlNodeType.Element || 
                                                       (n.NodeType == HtmlNodeType.Text && !string.IsNullOrWhiteSpace(n.InnerText)));
            var childList = children.ToList();
            for (int i = 0; i < childList.Count; i++)
            {
                var child = childList[i];
                bool isLastChild = i == childList.Count - 1;
                string newPrefix = prefix + (last ? "    " : "│   ");
                if (child.NodeType == HtmlNodeType.Element)
                {
                    RenderTree(child, newPrefix, isLastChild, showAttrs, selector);
                }
                else if (child.NodeType == HtmlNodeType.Text)
                {
                    string text = child.InnerText.Trim();
                    if (!string.IsNullOrEmpty(text))
                    {
                        string conn = isLastChild ? "└── " : "├── ";
                        Console.WriteLine($"{newPrefix}{conn}{Colorize($"\"{text}\"", "text")}");
                    }
                }
            }
        }
        else
        {
            // Комментарии и т.д.
            if (node.NodeType == HtmlNodeType.Comment)
            {
                string connector = last ? "└── " : "├── ";
                Console.WriteLine($"{prefix}{connector}{Colorize($"<!-- {node.InnerText} -->", "comment")}");
            }
        }
    }

    static List<string> ValidateHtml(HtmlDocument doc)
    {
        var errors = new List<string>();
        var ids = new HashSet<string>();
        var nodesWithId = doc.DocumentNode.SelectNodes("//*[@id]");
        if (nodesWithId != null)
        {
            foreach (var node in nodesWithId)
            {
                string id = node.GetAttributeValue("id", "");
                if (!string.IsNullOrEmpty(id))
                {
                    if (!ids.Add(id))
                        errors.Add($"Duplicate id '{id}'");
                }
            }
        }
        return errors;
    }

    static string Colorize(string text, string type)
    {
        if (!colorEnabled) return text;
        string code = type switch
        {
            "tag" => "\x1b[96m",
            "attr" => "\x1b[93m",
            "text" => "\x1b[92m",
            "comment" => "\x1b[90m",
            _ => "\x1b[0m"
        };
        return code + text + "\x1b[0m";
    }
}
