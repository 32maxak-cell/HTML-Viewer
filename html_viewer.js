// html_viewer.js
#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const cheerio = require('cheerio');
const chalk = require('chalk');

function colorize(text, type) {
    switch (type) {
        case 'tag': return chalk.cyan(text);
        case 'attr': return chalk.yellow(text);
        case 'text': return chalk.green(text);
        case 'comment': return chalk.gray(text);
        default: return text;
    }
}

function renderTree($, element, prefix = '', last = true, showAttrs = false, filterSel = null) {
    if (!element) return;
    const tagName = element[0].name;
    const id = element.attr('id') || '';
    const classes = (element.attr('class') || '').split(/\s+/).filter(Boolean);
    let display = tagName;
    if (id) display += `#${id}`;
    if (classes.length) display += `.${classes.join('.')}`;
    let attrStr = '';
    if (showAttrs) {
        const attrs = Object.entries(element.attr())
            .filter(([k]) => !['id', 'class'].includes(k))
            .map(([k, v]) => `${k}=${v}`);
        if (attrs.length) attrStr = ' ' + attrs.join(' ');
    }
    const connector = last ? '└── ' : '├── ';
    console.log(`${prefix}${connector}${colorize(display, 'tag')}${colorize(attrStr, 'attr')}`);
    
    const children = element.children();
    const childNodes = [];
    children.each((i, child) => {
        if (child.type === 'tag') {
            const childEl = $(child);
            if (filterSel) {
                // проверяем, соответствует ли элемент или его потомки селектору
                if (childEl.is(filterSel) || childEl.find(filterSel).length > 0) {
                    childNodes.push(childEl);
                }
            } else {
                childNodes.push(childEl);
            }
        } else if (child.type === 'text' && child.data.trim()) {
            childNodes.push(child.data.trim());
        }
    });
    const newPrefix = prefix + (last ? '    ' : '│   ');
    childNodes.forEach((node, idx) => {
        const isLast = idx === childNodes.length - 1;
        if (typeof node === 'string') {
            const connectorChild = isLast ? '└── ' : '├── ';
            console.log(`${newPrefix}${connectorChild}${colorize(`"${node}"`, 'text')}`);
        } else {
            renderTree($, node, newPrefix, isLast, showAttrs, filterSel);
        }
    });
}

function validateHTML($) {
    const errors = [];
    const ids = new Set();
    $('[id]').each((i, el) => {
        const id = $(el).attr('id');
        if (ids.has(id)) {
            errors.push(`Duplicate id '${id}'`);
        } else {
            ids.add(id);
        }
    });
    return errors;
}

function main() {
    const args = process.argv.slice(2);
    let tree = false;
    let color = false;
    let attributes = false;
    let validate = false;
    let selector = null;
    let output = null;
    let filePath = null;

    for (let i = 0; i < args.length; i++) {
        const arg = args[i];
        if (arg === '-t' || arg === '--tree') {
            tree = true;
        } else if (arg === '-c' || arg === '--color') {
            color = true;
        } else if (arg === '-a' || arg === '--attributes') {
            attributes = true;
        } else if (arg === '-v' || arg === '--validate') {
            validate = true;
        } else if (arg === '-s' || arg === '--selector') {
            if (i + 1 < args.length) {
                selector = args[++i];
            } else {
                console.error('Missing selector');
                process.exit(1);
            }
        } else if (arg === '-o' || arg === '--output') {
            if (i + 1 < args.length) {
                output = args[++i];
            } else {
                console.error('Missing output file');
                process.exit(1);
            }
        } else if (arg === '-h' || arg === '--help') {
            console.log(`Usage: node ${path.basename(__filename)} [options] [file]`);
            console.log('  -t, --tree          Show DOM tree');
            console.log('  -c, --color         Force color output');
            console.log('  -a, --attributes    Show attributes');
            console.log('  -v, --validate      Validate HTML');
            console.log('  -s, --selector      CSS selector filter');
            console.log('  -o, --output        Export to file');
            process.exit(0);
        } else {
            if (!filePath) filePath = arg;
            else { console.error(`Extra argument: ${arg}`); process.exit(1); }
        }
    }

    const useColor = color || process.stdout.isTTY;
    // Принудительно включаем chalk, если нужно
    if (!useColor) {
        chalk.level = 0;
    }

    let content;
    if (filePath) {
        try {
            content = fs.readFileSync(filePath, 'utf8');
        } catch (err) {
            console.error(`Error reading file: ${err.message}`);
            process.exit(1);
        }
    } else {
        if (process.stdin.isTTY) {
            console.error('No input provided. Pipe HTML or pass file.');
            process.exit(1);
        }
        content = fs.readFileSync(0, 'utf8');
    }

    const $ = cheerio.load(content);

    if (validate) {
        const errors = validateHTML($);
        if (errors.length === 0) {
            console.log(chalk.green('✅ No validation errors found.'));
        } else {
            console.log('Validation errors:');
            errors.forEach(err => console.log(chalk.red(`  ❌ ${err}`)));
        }
    }

    if (tree) {
        console.log(chalk.cyan('🌳 DOM Tree:'));
        const root = $('html');
        if (root.length === 0) {
            // fallback
            renderTree($, $('body').length ? $('body') : $('*').first(), '', true, attributes, selector);
        } else {
            renderTree($, root, '', true, attributes, selector);
        }
    }

    if (output) {
        const ext = output.split('.').pop().toLowerCase();
        if (ext === 'json') {
            // Простой экспорт: записываем структуру
            function serialize(el) {
                const obj = { name: el[0].name };
                const attrs = el.attr();
                if (Object.keys(attrs).length) obj.attrs = attrs;
                const children = [];
                el.children().each((i, child) => {
                    if (child.type === 'tag') {
                        children.push(serialize($(child)));
                    } else if (child.type === 'text' && child.data.trim()) {
                        children.push(child.data.trim());
                    }
                });
                if (children.length) obj.children = children;
                return obj;
            }
            const data = serialize($('html').length ? $('html') : $('*').first());
            fs.writeFileSync(output, JSON.stringify(data, null, 2));
            console.log(`Exported JSON to ${output}`);
        } else {
            // HTML
            fs.writeFileSync(output, $.html());
            console.log(`Exported HTML to ${output}`);
        }
    }
}

if (require.main === module) {
    main();
}
