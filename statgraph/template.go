// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package statgraph

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>Server Statistics &middot; httpstat</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">

	<style type="text/css">
		* { font-family: "Helvetica Neue", Helvetica, Arial, sans-serif; }
	</style>
</head>
<body>
	<h5>
		Request Per Second
		[<a href="./rps.png?w=800&h=200">png</a>]
		[<a href="./rps.svg?w=800&h=200">svg</a>]
		[<a href="./rps.png?w=800&h=200&spark=1">spark</a>]
	</h5>
	<img src="./rps.svg?w=800&h=200" id="rps_count">

	<h5>
		Request Count
		[<a href="./requests.png?w=800&h=200">png</a>]
		[<a href="./requests.svg?w=800&h=200">svg</a>]
		[<a href="./requests.png?w=800&h=200&spark=1">spark</a>]
	</h5>
	<img src="./requests.svg?w=800&h=200" id="request_count">

	<h5>
		Request Latency
		[<a href="./latency.png?w=800&h=200">png</a>]
		[<a href="./latency.svg?w=800&h=200">svg</a>]
		[<a href="./latency.png?w=800&h=200&spark=1">spark</a>]
	</h5>
	<img src="./latency.svg?w=800&h=200" id="request_latency">

	<script type="text/javascript">
		function timestamp() {
			return Math.round((new Date()).getTime() / 1000);
		}
		setInterval(function() {
			var rpsCount = document.getElementById('rps_count');
			rpsCount.src = './rps.svg?w=800&h=200&r=' + timestamp();

			var reqCount = document.getElementById('request_count');
			reqCount.src = './requests.svg?w=800&h=200&r=' + timestamp();

			var reqLatency = document.getElementById('request_latency');
			reqLatency.src = './latency.svg?w=800&h=200&r=' + timestamp();
		}, %d);
	</script>
</body>
</html>`
