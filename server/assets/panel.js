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
	// Create each row of the table (each key+value pair)
	var rows = [];
	for (var key in dict) {
		if (rows.indexOf(key) === -1) {
			rows.push(key);
		}
	}

	// Create and style the table
	var table = document.createElement("table");

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
				var wraiths = JSON.parse(wdata[1]["data"]);
				table_create("wraiths_page_table_container", wraiths);
				// Update list of command targets
				var dropdown = document.getElementById("console_input_target_selector");
				// Get the currently selected element to avoid removing the selection
				var selection = dropdown.value
				// Clear the dropdown to avoid keeping outdated wraiths or constantly re-adding existing ones
				dropdown.innerHTML = "";
				var count = 0;
				for (key in wraiths) {
					count += 1;
					// Skip the first entry because it's column headers
					if (count == 1) {continue;}
					// Append the element to the end of Array list
					dropdown[dropdown.length] = new Option(key, key);
				}
				// Set the selection to the old one
				dropdown.value = selection;
			} else if (wdata[0] == "settings") {
				table_create("settings_page_table_container", JSON.parse(wdata[1]["data"]));
			}
		}
	} else {
		alert("Your browser does not support workers. The page will need to be manually refreshed to update information.");
	}
}

