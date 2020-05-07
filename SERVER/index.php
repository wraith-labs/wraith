<!-- Minimal redirect page -->

<?php

$redirect_location = "PANEL/login.html";

// Redirect with a header
header("Location: " . $redirect_location);
?>

<head>

<!-- Favicon -->
<link rel="shortcut icon" href="favicon.png">

<script>
// Redirect with JavaScript
window.location.href = "<?php echo $redirect_location ?>";
</script>

</head>
<body>
Redirecting to "<?php echo $redirect_location ?>"...
</body>
