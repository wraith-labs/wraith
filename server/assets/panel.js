function api(args) {
	args["panel_token"] = current_panel_login_token;

	const Http = new XMLHttpRequest();
	const api_url = "api.php";
	const key = current_panel_crypt_key;
	const args_json = JSON.stringify(args);
	
	const crypt_prefix_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	var crypt_prefix = "";
	for (let i = 0; i < 5; i++) {
		crypt_prefix += crypt_prefix_charset[Math.floor(Math.random() * crypt_prefix_charset.length)];
	}
	crypt_prefix += "Wr7H"
	
	var response = "";
	var response_dict = {};
	
	Http.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			// The response could be encrypted but it might not be
			try {
		 		response = aes.decrypt(Http.responseText, key);
			 	// Parse the JSON response
			 	response_dict = JSON.parse(response);
		 		// Verify the fingerprint of the server (only if the communication is encrypted because otherwise the server does not attach it)
		 		if (response_dict["server_id"] != trusted_server_signature) {
		 			const untrusted_server_error = "The server provided an incorrect ID. This should never happen unless something went very wrong. The server ID is `" + response_dict["server_id"] + "` while the expected ID is `" + trusted_server_signature + "`. Quitting."
		 			throw new Error(untrusted_server_error);
		 		}
		 	} catch (err) {
		 		if (err.message.startsWith("The server provided an incorrect ID. ")) { throw err; }
		 		response = Http.responseText;
		 		// Parse the JSON response
		 		response_dict = JSON.parse(response);
		 	}
		 	// If the API returns a message, tell the user
		 	if ("message" in response_dict) {
		 		alert("The server says: "+response_dict["message"]);
		 	}
		// If the API returns an error in the form of a HTTP code
		} else if (this.readyState == 4 && this.status != 200) {
			alert("The API returned a non-200 code when called. The code was: " + Http.status);
		}
	}
	
	Http.open("POST", api_url, false);
	Http.send(crypt_prefix+aes.encrypt(args_json, key));

	return response_dict;
}

function table_create(parent_div_id, json_string) {
		var json_data = JSON.parse(json_string);
		var rows = [];
		for (var key in json_data) {
			if (rows.indexOf(key) === -1) {
				rows.push(key);
			}
		}

		var table = document.createElement("table");
		table.cellPadding = 10;
		table.style.border = "1px solid #aaaaaa";

		for (var i = 0; i < rows.length; i++) {
			var tr = table.insertRow(i);
			var th = document.createElement("th");
			var td = document.createElement("td");
			th.innerText = rows[i];
			td.innerText = json_data[rows[i]];
			th.style.border = "1px solid #aaaaaa";
			td.style.border = "1px solid #aaaaaa";
			tr.appendChild(th);
			tr.appendChild(td);
		}

		for (var i = 0; i < json_data.length; i++) {

			tr = table.insertRow(-1);

			for (var j = 0; j < col.length; j++) {
				var tabCell = tr.insertCell(-1);
				tabCell.innerHTML = json_data[i][col[j]];
			}
		}

		var divContainer = document.getElementById(parent_div_id);
		divContainer.innerHTML = "";
		divContainer.appendChild(table);
}

var page_update_worker;
function start_page_update_worker() {
	if (typeof(Worker) !== "undefined") {
		if (typeof(page_update_worker) == "undefined") {
			page_update_worker = new Worker("assets/panel_page_update_worker.js");
		}
		const baseuri_finder = new RegExp(/^.*\//);
		const base_uri = baseuri_finder.exec(window.location.href);
		page_update_worker.postMessage({"current_panel_login_token": current_panel_login_token, "current_panel_crypt_key": current_panel_crypt_key, "trusted_server_signature": trusted_server_signature, "base_uri": base_uri});
		page_update_worker.onmessage = function(event) {
			const wdata = event.data;
			if (wdata[0] == "info") {
				table_create("info_page_table_container", wdata[1]["data"]);
			} else if (wdata[0] == "wraiths") {
				table_create("wraiths_page_table_container", wdata[1]["data"]);
			}
		}
	} else {
		alert("Your browser does not support workers. The page will need to be manually refreshed to refresh information.");
	}
}

