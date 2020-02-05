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

// Function which sends the command in the command input box to the selected wraith(s)
function sendCommand() {
	const target = document.getElementById("console_input_target_selector").value;
	const command = document.getElementById("console_input_command_entry").value;
	
	// Check the command and target
	if (target == "") {
		alert("Please select a target to send the command to!");
	} else if (command == "") {
		alert("Please enter a command to send!");
	} else {
		
		const data = {
			"targets": [target], // TODO: Allow selecting multiple targets
			"command": command
		};

		// Clear the command window to show the command was sent
		document.getElementById("console_input_command_entry").value = "";

		api({"message_type": "sendcommand", "data": data});
	
	}

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
				table_create("info_page_table_container", JSON.parse(wdata[1]));

			} else if (wdata[0] == "wraiths") {
				var wraiths = JSON.parse(wdata[1]);
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
					// Replace it with the "all" selector
					if (count == 1) {key = "All";}
					// Append the element to the end of Array list
					dropdown[dropdown.length] = new Option(key, key);
				}
				// Set the selection to the old one
				dropdown.value = selection;

			} else if (wdata[0] == "console") {
				// Get the console output section
				var console_out = document.getElementById("console_output_container");
				// Clear the console output section
				console_out.innerHTML = "";
				// Populate console output section with contents of array
				const data = JSON.parse(wdata[1]);
				// Define node as a variable here so it can be later accessed outside of the loop for scrolling purposes
				var node
				for (i = 0; i < data.length; ++i) {
					node = document.createElement("li");
					// Output formatting ([ src/dst ] < timestamp > ( status ) text)
					node.innerText = "[ "+data[i][0]+" ]  < "+new Date(data[i][1] * 1000).toISOString().slice(0, 19).replace('T', ' ')+" >  ( "+data[i][2]+" ) | "+data[i][3];
					// Add line to output box
					console_out.appendChild(node);
				}
				// Scroll to the newest line (comment out to stop this behaviour)
				console_out.scrollTop = node.offsetTop
				
			} else if (wdata[0] == "settings") {
				table_create("settings_page_table_container", JSON.parse(wdata[1]));
			}
		}
	} else {
		alert("Your browser does not support workers. The page will need to be manually refreshed to update information.");
	}
}

