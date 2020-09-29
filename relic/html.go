package relic

const inlineTemplate = `
<!DOCTYPE html>
<html>

<head>
    <style>
        body {
            margin: 20;
            font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
            font-size: 14px;
            line-height: 1.4;
            color: #222;
        }

        a {
            color: #333;
        }

        .container {
            padding: 10px;
        }

        .subtext {
            font-size: 11px;
            font-style: italic;
            margin-top: 0px;
            margin-bottom: 10px;
            color: #888;
        }
    </style>
</head>

<body>
    {{if .recent}}
    <div class="container">
        {{ .recent }}
        <div class="subtext">{{ .recentdate }}</div>
    </div>
    {{end}}
    <div class="container">
        {{ .random }}
        <div class="subtext">{{ .randomdate }}</div>
    </div>
    <div class="container">
        <div class="subtext"><a href="https://pinboard.in">Go to Pinboard</a></div>
    </div>
</body>

</html>
`
