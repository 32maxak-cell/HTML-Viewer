# html_viewer.py
import sys
import argparse
import json
from bs4 import BeautifulSoup, Tag, NavigableString, Comment
import re
from collections import defaultdict

# ANSI colors
COLORS = {
    'tag': '\033[96m',       # cyan
    'attr': '\033[93m',      # yellow
    'text': '\033[92m',      # green
    'comment': '\033[90m',   # grey
    'reset': '\033[0m'
}

def colorize(text, color, enabled):
    return f"{COLORS[color]}{text}{COLORS['reset']}" if enabled else text

def parse_html(content):
    return BeautifulSoup(content, 'lxml')

def validate_html(soup):
    errors = []
    # Проверка незакрытых тегов (BeautifulSoup может их додумывать)
    # Проверка дублирующихся id
    ids = {}
    for tag in soup.find_all(True):
        if tag.get('id'):
            id_val = tag['id']
            if id_val in ids:
                errors.append(f"Duplicate id '{id_val}' found in <{tag.name}>")
            else:
                ids[id_val] = tag.name
    # Проверка вложенности (можно добавить больше)
    return errors

def render_tree(element, prefix="", last=True, color=True, show_attrs=False, selector=None):
    if selector and not element.select_one(selector):
        # Если фильтр задан, пропускаем элементы, не подходящие, но их потомки могут подходить
        # Просто рекурсивно обрабатываем потомков
        if hasattr(element, 'children'):
            children = list(element.children)
            for i, child in enumerate(children):
                if isinstance(child, Tag):
                    render_tree(child, prefix + ("    " if last else "│   "), i == len(children)-1, color, show_attrs, selector)
        return

    if isinstance(element, Tag):
        # Вывод тега
        connector = "└── " if last else "├── "
        tag_name = element.name
        id_attr = element.get('id')
        classes = element.get('class', [])
        tag_str = tag_name
        if id_attr:
            tag_str += f"#{id_attr}"
        if classes:
            tag_str += "." + ".".join(classes)
        attr_str = ""
        if show_attrs:
            attrs = {k: v for k, v in element.attrs.items() if k not in ('id', 'class')}
            if attrs:
                attr_str = " " + " ".join(f"{k}={v}" for k, v in attrs.items())
        print(f"{prefix}{connector}{colorize(tag_str, 'tag', color)}{colorize(attr_str, 'attr', color)}")
        
        # Рекурсивно обрабатываем детей
        children = [c for c in element.children if isinstance(c, (Tag, NavigableString))]
        for i, child in enumerate(children):
            new_prefix = prefix + ("    " if last else "│   ")
            if isinstance(child, Tag):
                render_tree(child, new_prefix, i == len(children)-1, color, show_attrs, selector)
            elif isinstance(child, NavigableString) and str(child).strip():
                text = str(child).strip()
                if text:
                    print(f"{new_prefix}{'└── ' if i == len(children)-1 else '├── '}{colorize(f'\"{text}\"', 'text', color)}")
    else:
        # Комментарии и т.п.
        if isinstance(element, Comment):
            print(f"{prefix}{'└── ' if last else '├── '}{colorize('<!-- ' + str(element) + ' -->', 'comment', color)}")

def main():
    parser = argparse.ArgumentParser(description="HTML Viewer (Developer Mode)")
    parser.add_argument('file', nargs='?', help='HTML file to process')
    parser.add_argument('-t', '--tree', action='store_true', help='Show DOM tree')
    parser.add_argument('-c', '--color', action='store_true', help='Force color output')
    parser.add_argument('-s', '--selector', help='CSS selector filter')
    parser.add_argument('-a', '--attributes', action='store_true', help='Show attributes')
    parser.add_argument('-v', '--validate', action='store_true', help='Validate HTML')
    parser.add_argument('-o', '--output', help='Export to file (JSON or HTML)')
    args = parser.parse_args()

    color = args.color or sys.stdout.isatty()

    # Чтение данных
    if args.file:
        with open(args.file, 'r', encoding='utf-8') as f:
            content = f.read()
    else:
        if sys.stdin.isatty():
            print("No input provided. Pipe HTML or pass file.", file=sys.stderr)
            sys.exit(1)
        content = sys.stdin.read()

    soup = parse_html(content)

    # Валидация
    if args.validate:
        errors = validate_html(soup)
        if errors:
            print("Validation errors:")
            for err in errors:
                print(colorize(f"  ❌ {err}", 'tag', color))
        else:
            print(colorize("✅ No validation errors found.", 'text', color))

    # Дерево
    if args.tree:
        print(colorize("🌳 DOM Tree:", 'tag', color))
        # Ищем корневой элемент (обычно html)
        root = soup.find('html')
        if not root:
            root = soup  # fallback
        render_tree(root, color=color, show_attrs=args.attributes, selector=args.selector)

    # Экспорт
    if args.output:
        ext = args.output.split('.')[-1].lower()
        if ext == 'json':
            # Экспорт в JSON (просто структура)
            def node_to_dict(tag):
                if isinstance(tag, Tag):
                    return {
                        'name': tag.name,
                        'attrs': tag.attrs,
                        'children': [node_to_dict(c) for c in tag.children if isinstance(c, (Tag, NavigableString)) and (not isinstance(c, NavigableString) or str(c).strip())]
                    }
                elif isinstance(tag, NavigableString):
                    return str(tag).strip()
                return None
            data = node_to_dict(soup)
            with open(args.output, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
            print(f"Exported JSON to {args.output}")
        elif ext == 'html':
            with open(args.output, 'w', encoding='utf-8') as f:
                f.write(soup.prettify())
            print(f"Exported formatted HTML to {args.output}")
        else:
            print("Unsupported export format. Use .json or .html", file=sys.stderr)

if __name__ == '__main__':
    main()
