🔍 HTML Viewer – Режим разработчика
Мощный консольный инструмент для анализа, отладки и визуализации HTML-документов
Поддерживает 7 языков программирования – выбирайте свой любимый!

✨ Возможности
🌳 Просмотр DOM-дерева – наглядное представление структуры документа с отступами и индикаторами типов узлов.

🎨 Цветная подсветка – теги, атрибуты, текст и комментарии выделены ANSI-цветами для удобства чтения.

🔍 Фильтрация по CSS-селектору – показывать только элементы, соответствующие запросу (например, div.container > p).

🧩 Показ атрибутов – вывод всех атрибутов элемента (id, class, data-*, и т.д.).

📐 Стили – отображение вычисленных стилей (инлайн-стили и, при наличии, внешние CSS).

❌ Валидация разметки – поиск незакрытых тегов, дублирующихся id, недопустимых атрибутов.

💾 Экспорт – сохранение DOM-дерева в формате JSON или переформатированного HTML.

⚡ Работа из командной строки – читает файл или стандартный ввод.

📦 Поддерживаемые языки
Язык	Версия	Файл	Основная библиотека
Python	3.8+	html_viewer.py	beautifulsoup4, lxml
Go	1.18+	html_viewer.go	golang.org/x/net/html
Rust	1.60+	html_viewer.rs	scraper
JavaScript	Node.js 14+	html_viewer.js	cheerio / jsdom
C#	.NET 6+	html_viewer.cs	HtmlAgilityPack
Java	11+	HtmlViewer.java	jsoup
C++	C++17	html_viewer.cpp	gumbo-parser
🚀 Быстрый старт
1. Склонируйте репозиторий
bash
git clone https://github.com/yourname/html-viewer.git
cd html-viewer
2. Запустите на любом языке
Python

bash
pip install beautifulsoup4 lxml
python html_viewer.py index.html -t -c -s ".header"
Go

bash
go mod init html_viewer
go get golang.org/x/net/html
go run html_viewer.go index.html -t -c -s ".header"
Rust (сборка)

bash
cargo new html_viewer
# добавьте зависимости в Cargo.toml
cargo run -- index.html -t -c -s ".header"
JavaScript (Node.js)

bash
npm install cheerio chalk
node html_viewer.js index.html -t -c -s ".header"
C#

bash
dotnet new console -n html_viewer
dotnet add package HtmlAgilityPack
dotnet run -- index.html -t -c -s ".header"
Java (сборка с Maven/Gradle)

bash
javac -cp .:jsoup.jar HtmlViewer.java
java -cp .:jsoup.jar HtmlViewer index.html -t -c -s ".header"
C++ (сборка с gumbo-parser)

bash
g++ -std=c++17 -I/usr/include/gumbo html_viewer.cpp -lgumbo -o html_viewer
./html_viewer index.html -t -c -s ".header"
📋 Пример вывода
Для файла index.html:

html
<!DOCTYPE html>
<html>
<head><title>Hello</title></head>
<body>
  <div id="main" class="container">
    <p>Text here</p>
  </div>
</body>
</html>
Программа с опцией -t -c выдаст:

text
🌳 DOM Tree:
📄 html (2 children)
 ├── 📄 head (1 child)
 │   └── 📄 title (1 child)
 │       └── 📝 "Hello"
 └── 📄 body (1 child)
     └── 📄 div#main.container (1 child)
         ├── 📄 p (1 child)
         │   └── 📝 "Text here"
         └── (attr) id="main"
         └── (attr) class="container"
С фильтром -s "#main" будет показан только элемент div#main и его потомки.

⚙️ Опции командной строки
Флаг	Описание
-t, --tree	Показать DOM-дерево (по умолчанию только валидация)
-c, --color	Включить цветной вывод
-s, --selector <expr>	CSS-селектор для фильтрации элементов
-a, --attributes	Показывать атрибуты для каждого элемента
-v, --validate	Проверить валидность разметки (незакрытые теги, дублирующиеся id)
-o, --output <file>	Экспортировать дерево в JSON или HTML (зависит от расширения)
-h, --help	Справка
📄 Лицензия
MIT – свободно используйте, модифицируйте и распространяйте.

🤝 Вклад
Приветствуются pull request'ы! Если хотите добавить новый язык или улучшить существующий – создавайте issue.

🧠 Авторы
Проект создан в образовательных целях для демонстрации парсинга HTML и анализа на разных языках.

