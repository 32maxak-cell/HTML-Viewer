// html_viewer.rs
use std::env;
use std::fs;
use std::io::{self, Read};
use std::process;
use scraper::{Html, Selector, ElementRef, Node};
use colored::*;

fn colorize(text: &str, color: &str) -> String {
    match color {
        "tag" => text.cyan().to_string(),
        "attr" => text.yellow().to_string(),
        "text" => text.green().to_string(),
        "comment" => text.bright_black().to_string(),
        _ => text.to_string(),
    }
}

fn render_tree(element: ElementRef, prefix: &str, last: bool, color: bool, show_attrs: bool, filter_sel: Option<&Selector>) {
    let connector = if last { "└── " } else { "├── " };
    let tag_name = element.value().name();
    let id = element.value().attr("id").unwrap_or("");
    let classes = element.value().attr("class").unwrap_or("");
    let mut display = tag_name.to_string();
    if !id.is_empty() {
        display.push_str(&format!("#{}", id));
    }
    if !classes.is_empty() {
        let class_list: Vec<&str> = classes.split_whitespace().collect();
        if !class_list.is_empty() {
            display.push_str(".");
            display.push_str(&class_list.join("."));
        }
    }
    let attr_str = if show_attrs {
        let mut attrs = Vec::new();
        for attr in element.value().attrs() {
            if attr.0 != "id" && attr.0 != "class" {
                attrs.push(format!("{}={}", attr.0, attr.1));
            }
        }
        if !attrs.is_empty() {
            format!(" {}", attrs.join(" "))
        } else {
            String::new()
        }
    } else {
        String::new()
    };

    if color {
        println!("{}{}{}{}", prefix, connector, colorize(&display, "tag"), colorize(&attr_str, "attr"));
    } else {
        println!("{}{}{}{}", prefix, connector, display, attr_str);
    }

    let children: Vec<_> = element.children().filter_map(ElementRef::wrap).collect();
    let text_nodes: Vec<_> = element.children().filter(|n| n.value().is_text()).collect();
    let mut all_nodes = Vec::new();
    // Interleave text and element nodes
    for child in element.children() {
        if let Some(el) = ElementRef::wrap(child) {
            all_nodes.push(el);
        } else if let Some(text) = child.value().as_text() {
            let trimmed = text.trim();
            if !trimmed.is_empty() {
                all_nodes.push(trimmed);
            }
        }
    }

    for (i, node) in all_nodes.iter().enumerate() {
        let is_last = i == all_nodes.len() - 1;
        let new_prefix = format!("{}{}", prefix, if last { "    " } else { "│   " });
        if let Some(el) = node.downcast_ref::<ElementRef>() {
            // Проверка фильтра
            if let Some(sel) = filter_sel {
                if !el.select(sel).next().is_some() && !el.value().parent().and_then(|p| ElementRef::wrap(p)).map(|p| p.select(sel).next().is_some()).unwrap_or(false) {
                    continue;
                }
            }
            render_tree(*el, &new_prefix, is_last, color, show_attrs, filter_sel);
        } else if let Some(text) = node.downcast_ref::<&str>() {
            let connector_child = if is_last { "└── " } else { "├── " };
            let quoted = format!("\"{}\"", text);
            if color {
                println!("{}{}{}", new_prefix, connector_child, colorize(&quoted, "text"));
            } else {
                println!("{}{}{}", new_prefix, connector_child, quoted);
            }
        }
    }
}

fn validate_html(document: &Html) -> Vec<String> {
    let mut errors = Vec::new();
    let mut ids = std::collections::HashSet::new();
    let selector = Selector::parse("[id]").unwrap();
    for element in document.select(&selector) {
        if let Some(id) = element.value().attr("id") {
            if !ids.insert(id) {
                errors.push(format!("Duplicate id '{}'", id));
            }
        }
    }
    // Другие проверки можно добавить
    errors
}

fn main() {
    let args: Vec<String> = env::args().collect();
    let mut tree = false;
    let mut color = false;
    let mut attributes = false;
    let mut validate = false;
    let mut selector = None;
    let mut output = None;
    let mut file_path = None;

    let mut i = 1;
    while i < args.len() {
        match args[i].as_str() {
            "-t" | "--tree" => { tree = true; i += 1; }
            "-c" | "--color" => { color = true; i += 1; }
            "-a" | "--attributes" => { attributes = true; i += 1; }
            "-v" | "--validate" => { validate = true; i += 1; }
            "-s" | "--selector" => {
                if i + 1 < args.len() {
                    selector = Some(args[i+1].clone());
                    i += 2;
                } else {
                    eprintln!("Missing selector");
                    process::exit(1);
                }
            }
            "-o" | "--output" => {
                if i + 1 < args.len() {
                    output = Some(args[i+1].clone());
                    i += 2;
                } else {
                    eprintln!("Missing output file");
                    process::exit(1);
                }
            }
            "-h" | "--help" => {
                println!("Usage: {} [options] [file]", args[0]);
                println!("  -t, --tree          Show DOM tree");
                println!("  -c, --color         Force color output");
                println!("  -a, --attributes    Show attributes");
                println!("  -v, --validate      Validate HTML");
                println!("  -s, --selector      CSS selector filter");
                println!("  -o, --output        Export to file");
                process::exit(0);
            }
            _ => {
                if file_path.is_none() {
                    file_path = Some(args[i].clone());
                    i += 1;
                } else {
                    eprintln!("Extra argument: {}", args[i]);
                    process::exit(1);
                }
            }
        }
    }

    let use_color = color || atty::is(atty::Stream::Stdout);

    let content = if let Some(path) = file_path {
        fs::read_to_string(&path).unwrap_or_else(|e| {
            eprintln!("Error reading file: {}", e);
            process::exit(1);
        })
    } else {
        let mut buffer = String::new();
        if io::stdin().read_to_string(&mut buffer).is_err() || buffer.is_empty() {
            eprintln!("No input provided. Pipe HTML or pass file.");
            process::exit(1);
        }
        buffer
    };

    let document = Html::parse_document(&content);

    if validate {
        let errors = validate_html(&document);
        if errors.is_empty() {
            println!("{}", if use_color { "✅ No validation errors found.".green() } else { "✅ No validation errors found." });
        } else {
            println!("Validation errors:");
            for err in errors {
                let colored = if use_color { format!("  ❌ {}", err).red() } else { format!("  ❌ {}", err) };
                println!("{}", colored);
            }
        }
    }

    if tree {
        let root = document.root_element();
        let filter_sel = selector.as_ref().and_then(|s| Selector::parse(s).ok());
        println!("{}", if use_color { "🌳 DOM Tree:".cyan() } else { "🌳 DOM Tree:" });
        render_tree(root, "", true, use_color, attributes, filter_sel.as_ref());
    }

    if let Some(out) = output {
        // Просто сохраняем оригинальный HTML (или можно форматированный)
        if let Err(e) = fs::write(&out, &content) {
            eprintln!("Error writing output: {}", e);
        } else {
            println!("Exported to {}", out);
        }
    }
}
