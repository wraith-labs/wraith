<?php
session_start();

require_once("assets/functions.php");

$db = get_db();

$failedlogin = false;

if (isset($_POST["LOGMEINPLZ"])) {
	if ($_POST["USER"] == $db["username"] && $_POST["PASS"] == $db["password"]) {
		$_SESSION["LOGGED_IN"] = true;
		$_SESSION["USERNAME"] = $_POST["USER"];
		$_SESSION["PASSWORD"] = $_POST["PASS"];
		header("Location: panel.php");
		die("SUCCESS");
	} else {
		$failedlogin = true;	
	}
	unset($_POST);
}

if (isset($_GET["LOGMEOUTPLZ"])) {
	if ($_SESSION["LOGGED_IN"] == true) {
		unset($_SESSION["LOGGED_IN"]);
		unset($_SESSION["USERNAME"]);
		unset($_SESSION["PASSWORD"]);
		header("Location: login.php");
		die("SUCCESS");
	} else {
		header("Location: login.php");
		die("Not logged in so not logged out!");
	}
	unset($_GET);
}

if ($_SESSION["LOGGED_IN"] == true && $_SESSION["USERNAME"] == $db["username"] && $_SESSION["PASSWORD"] == $db["password"]) {
	header("Location: panel.php");
	die("Logged in");
}

?>

<!--

Why are you looking here? Don't you have something
better to do? Surely you do? Right? Then go do it.
 :]

-->

<!DOCTYPE html>
<html>
	<head>
		<title>Wraith Login</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="assets/all.css" rel="stylesheet" type="text/css">
		<link href="assets/login.css" rel="stylesheet" type="text/css">
	</head>
	<body>
		<div id="login" class="login center">
			<h1>Wraith Login</h1>
			<?php if ($failedlogin) { ?><div style="background-color: red; margin: auto; padding: 3px 30px; display: inline-block; color: #eee;">Login Failed</div><?php } ?>
			<form id="login_form" action="" method="post">
				<input type="hidden" name="LOGMEINPLZ" value="true"></input>
				<input type="username" placeholder="Username" name="USER"></input>
				<input type="password" placeholder="Password" name="PASS"></input>
				<input type="submit" value="Log In"></input>
			</form>
		</div>
	</body>
</html>
