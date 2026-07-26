// HtmlViewer.java
import org.jsoup.Jsoup;
import org.jsoup.nodes.Document;
import org.jsoup.nodes.Element;
import org.jsoup.nodes.Node;
import org.jsoup.nodes.TextNode;
import org.jsoup.select.Elements;
import org.jsoup.select.NodeVisitor;

import java.io.*;
import java.nio.file.*;
import java.util.*;

public class HtmlViewer {
    private static boolean color;

    public static void main(String[] args) throws Exception {
        boolean tree = false;
        boolean forceColor = false;
        boolean attributes = false;
        boolean validate = false;
        String selector = null;
        String output = null;
        String filePath = null;

        for (int i = 0; i < args.length; i++) {
            switch (args[i]) {
                case "-t":
                case "--tree":
                    tree = true;
                    break;
                case "-c":
                case "--color":
                    forceColor = true;
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
                    if (i + 1 < args.length) selector = args[++i];
                    else { System.err.println("Missing selector"); System.exit(1); }
                    break;
                case "-o":
                case "--output":
                    if (i + 1 < args.length) output = args[++i];
                    else { System.err.println("Missing output file"); System.exit(1); }
                    break;
                case "-h":
                case "--help":
                    System.out.println("Usage: java HtmlViewer [options] [file]");
                    System.out.println("  -t, --tree          Show DOM tree");
                    System.out.println("  -c, --color         Force color output");
                    System.out.println("  -a, --attributes    Show attributes");
                    System.out.println("  -v, --validate      Validate HTML");
                    System.out.println("  -s, --selector      CSS selector filter");
                    System.out.println("  -o, --output        Export to file");
                    System.exit(0);
                    break;
                default:
                    if (filePath == null) filePath = args[i];
                    else { System.err.println("Extra argument: " + args[i]); System.exit(1); }
            }
        }

        color = forceColor || System.console() != null;

        String content;
        if (filePath != null) {
            content = new String(Files.readAllBytes(Paths.get(filePath)));
        } else {
            if (System.console() != null) {
                System.err.println("No input provided. Pipe HTML or pass file.");
                System.exit(1);
                return;
            }
            StringBuilder sb = new StringBuilder();
            try (BufferedReader br = new BufferedReader(new InputStreamReader(System.in))) {
                String line;
                while ((line = br.readLine()) != null) sb.append(line).append("\n");
            }
            content = sb.toString();
        }

        Document doc = Jsoup.parse(content);

        if (validate) {
            List<String> errors = validateHtml(doc);
            if (errors.isEmpty()) {
                System.out.println(colorize("✅ No validation errors found.", "text"));
            } else {
                System.out.println("Validation errors:");
                for (String err : errors) {
                    System.out.println(colorize("  ❌ " + err, "tag"));
                }
            }
        }

        if (tree) {
            System.out.println(colorize("🌳 DOM Tree:", "tag"));
            Element root = doc.select("html").first();
            if (root == null) root = doc.body(); // fallback
            renderTree(root, "", true, attributes, selector);
        }

        if (output != null) {
            Files.write(Paths.get(output), doc.html().getBytes());
            System.out.println("Exported to " + output);
        }
    }

    static void renderTree(Node node, String prefix, boolean last, boolean showAttrs, String selector) {
        if (node instanceof Element) {
            Element el = (Element) node;
            String connector = last ? "└── " : "├── ";
            String tagName = el.tagName();
            String id = el.id();
            String classNames = el.className();
            StringBuilder display = new StringBuilder(tagName);
            if (!id.isEmpty()) display.append("#").append(id);
            if (!classNames.isEmpty()) {
                String[] classes = classNames.split(" ");
                display.append(".").append(String.join(".", classes));
            }
            String attrStr = "";
            if (showAttrs) {
                List<String> attrs = new ArrayList<>();
                for (var attr : el.attributes()) {
                    if (!attr.getKey().equals("id") && !attr.getKey().equals("class")) {
                        attrs.add(attr.getKey() + "=" + attr.getValue());
                    }
                }
                if (!attrs.isEmpty()) attrStr = " " + String.join(" ", attrs);
            }

            System.out.println(prefix + connector + colorize(display.toString(), "tag") + colorize(attrStr, "attr"));

            // Проверка фильтра
            boolean matches = true;
            if (selector != null && !selector.isEmpty()) {
                Elements selected = el.select(selector);
                if (selected.isEmpty()) {
                    // Проверим, есть ли подходящие потомки
                    if (el.select(selector).isEmpty()) {
                        matches = false;
                    }
                }
            }

            if (matches) {
                List<Node> children = new ArrayList<>();
                for (Node child : el.childNodes()) {
                    if (child instanceof Element || (child instanceof TextNode && !((TextNode) child).text().trim().isEmpty())) {
                        children.add(child);
                    }
                }
                for (int i = 0; i < children.size(); i++) {
                    Node child = children.get(i);
                    boolean isLastChild = i == children.size() - 1;
                    String newPrefix = prefix + (last ? "    " : "│   ");
                    if (child instanceof Element) {
                        renderTree(child, newPrefix, isLastChild, showAttrs, selector);
                    } else if (child instanceof TextNode) {
                        String text = ((TextNode) child).text().trim();
                        if (!text.isEmpty()) {
                            String conn = isLastChild ? "└── " : "├── ";
                            System.out.println(newPrefix + conn + colorize("\"" + text + "\"", "text"));
                        }
                    }
                }
            }
        } else if (node instanceof TextNode) {
            // не обрабатываем отдельно
        }
    }

    static List<String> validateHtml(Document doc) {
        List<String> errors = new ArrayList<>();
        Set<String> ids = new HashSet<>();
        Elements elementsWithId = doc.select("[id]");
        for (Element el : elementsWithId) {
            String id = el.id();
            if (!id.isEmpty()) {
                if (!ids.add(id)) {
                    errors.add("Duplicate id '" + id + "'");
                }
            }
        }
        return errors;
    }

    static String colorize(String text, String type) {
        if (!color) return text;
        String code;
        switch (type) {
            case "tag": code = "\033[96m"; break;
            case "attr": code = "\033[93m"; break;
            case "text": code = "\033[92m"; break;
            case "comment": code = "\033[90m"; break;
            default: code = "\033[0m";
        }
        return code + text + "\033[0m";
    }
}
