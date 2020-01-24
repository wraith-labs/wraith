function api() {
	const Http = new XMLHttpRequest();
	const url='/api.php';
	Http.open("POST", url);
	Http.send();

	Http.onreadystatechange = (e) => {
	  return Http.responseText
	}
}
