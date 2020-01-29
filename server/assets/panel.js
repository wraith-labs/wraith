// Get the base URI of the webpage (useful when calling the API)
const baseuri_finder = new RegExp(/^.*\//);
const base_uri = baseuri_finder.exec(window.location.href);

// Switch to page with given ID
function switch_to_page(page) {
	// If the page is not already the page provided...
	if (window.location.hash != page) {
		// ...switch to that page
		window.location.hash = page;
	}
}

// A function which creates tables based on a dictionary
// Each key is a heading and each value is the value
function table_create(parent_div_id, dict) {
	// Create the style for the table
	const table_style = "1px solid #aaaaaa";

	// Create each row of the table (each key+value pair)
	var rows = [];
	for (var key in dict) {
		if (rows.indexOf(key) === -1) {
			rows.push(key);
		}
	}

	// Create and style the table
	var table = document.createElement("table");
	table.cellPadding = 10;
	table.style.border = table_style;

	// Create the cells with the key and value
	for (var i = 0; i < rows.length; i++) {
		// Create a row in the actual table
		var tr = table.insertRow(i);
		// Create the cells for key and value
		var th = document.createElement("th");
		var td = document.createElement("td");
		// Add text to the cells
		th.innerText = rows[i];
		td.innerText = dict[rows[i]];
		// Style the cells
		th.style.border = table_style;
		td.style.border = table_style;
		// Add the cells to the table
		tr.appendChild(th);
		tr.appendChild(td);
	}

	// Finally, get the DIV that holds the table
	var divContainer = document.getElementById(parent_div_id);
	// Clear it (to avoid filling the page with tables)
	divContainer.innerHTML = "";
	// Add the table to the div
	divContainer.appendChild(table);
}

var page_update_worker;
function start_page_update_worker() {
	if (typeof(Worker) !== "undefined") {
		if (typeof(page_update_worker) == "undefined") {
			page_update_worker = new Worker("assets/panel_page_update_worker.js");
		}
		page_update_worker.postMessage({"current_panel_login_token": current_panel_login_token, "current_panel_crypt_key": current_panel_crypt_key, "trusted_server_signature": trusted_server_signature, "base_uri": base_uri});
		page_update_worker.onmessage = function(event) {
			const wdata = event.data;
			if (wdata[0] == "info") {
				table_create("info_page_table_container", JSON.parse(wdata[1]["data"]));
			} else if (wdata[0] == "wraiths") {
				table_create("wraiths_page_table_container", JSON.parse(wdata[1]["data"]));
			} else if (wdata[0] == "settings") {
				table_create("settings_page_table_container", JSON.parse(wdata[1]["data"]));
			}
		}
	} else {
		alert("Your browser does not support workers. The page will need to be manually refreshed to update information.");
	}
}

