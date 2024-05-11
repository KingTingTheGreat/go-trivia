const connectWS = (endpoint, func) => {
	const socket = new WebSocket(endpoint);
	
	socket.onopen = () => { console.log(`${endpoint}: connected`)};

	socket.onmessage = (e) => { func(JSON.parse(e.data)) };

	socket.onclose = () => { setTimeout(() => {connectWS(endpoint), 100}) };
}
