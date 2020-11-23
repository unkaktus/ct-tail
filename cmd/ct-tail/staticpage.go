package main

var staticPage = `<html>
<head>
	<script>
        function printLine(line) {
            if (line == "") {
                return
            }
            today = new Date()
            date = today.toLocaleString()
            line = date + "\t" + line

            let elem = document.getElementById('log')
            let lineElem = document.createElement('pre');
            lineElem.append(line);
            elem.append(lineElem);

            lineElem.scrollIntoView()
        }

        function showEntry(entry) {
            let line = entry.identifiers.dns_names.filter(function (val) {return val;}).join(' ')
            printLine(line)
        }

		window.onload = function runLog() {
            ws = new WebSocket("wss://nogoegst.net/ct-tail/log/ws");
            ws.onopen = function(evt) {
                printLine("CT Log tailing started. DNS names for new log entries will appear below.")
            }
            ws.onclose = function(evt) {
                printLine("Stream ended.")
                return;
            }

            ws.onerror = function(evt) {
                printLine("Stream ended due to error.")
                return;
            }

            ws.onmessage = function(evt) {
                if (evt.data != "") {
                    entry = JSON.parse(evt.data)
                    showEntry(entry)
                }
            }
        }
	</script>
</head>
<body>
    <div id="log">
    </div>
</body>
</html>
`
