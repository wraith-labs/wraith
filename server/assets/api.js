function api(args) {
	const Http = new XMLHttpRequest();
	const api_url='/api.php';
	const key = current_panel_crypt_key;
	
	Http.open("POST", api_url, true);
	Http.send(aes.encrypt(args, key));

	Http.onreadystatechange = (e) => {
	  alert(Http.responseText);
	}
}
