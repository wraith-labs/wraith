<?php
// Start session for user to track login etc.
session_start();

// Include the functions required
include_once("assets/functions.php");

// Get saved database
$db = get_db();

// If the user is not logged in, redirect them and die without sending any HTML
if (!($_SESSION["LOGGED_IN"] == true && $_SESSION["USERNAME"] == $db["username"] && $_SESSION["PASS"] == $db["PASSWORD"])) {
	header("Location: login.php");
	die("Log in first!");
} else {
	panel_login();
}

// Request credentials from the API

?>

<!--

Why are you looking here? Don't you have something
better to do? Surely you do? Right? Then go do it.
:]

-->

<!DOCTYPE html>
<html>
	<head>
		<title>Wraith Panel</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="assets/all.css" rel="stylesheet" type="text/css">
		<link href="assets/panel.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="assets/panel.js"></script>
		<script type="text/javascript" src="assets/crypto.js"></script>
		<script type="text/javascript" src="assets/api.js"></script>
		<script type="text/javascript">
			<?php $db = get_db(); ?>
			const current_panel_login_token = "<?php echo $db['current_panel_login_token']; ?>";
			const current_panel_crypt_key = "<?php echo $db['current_panel_crypt_key']; ?>";
			const trusted_server_signature = "<?php echo $db['server_id']; ?>";
			
			// Update page every 3 seconds
			(function update_page() {
				// Get info for info page
				const info_response = api({"message_type": "getinfo"});
				// Get info about wraiths for wraiths page
				// Get a list of options for the options page

				setTimeout(update_page, 3000);
			})();
		</script>
	</head>
	<body>
		<div class="sidenav" id="sidenav">
			<h3>Wraith Panel</h3>
			<a href="#info_page">Info</a>
			<a href="#wraiths_page">Wraiths</a>
			<a href="#command_center_page">Command Center</a>
			<a href="#server_options_page">Server Options</a>
			<a href="login.php?LOGMEOUTPLZ=true">Log Out</a>
		</div>
		<div name="info_page" id="info_page" class="page">
			Info
		</div>
		<div name="wraiths_page" id="wraiths_page" class="page">
			Wraiths
		</div>
		<div name="command_center_page" id="command_center_page" class="page">
			Command Center
		</div>
		<div name="server_options_page" id="server_options_page" class="page">
			Settings
		</div>
	</body>
</html>
