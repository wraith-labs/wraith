<?php

/* This is the location of the API. It can be relative or
it can contain a complete URL if the API is hosted on a
separate server. If you have kept the folder structure
unchanged and the panel and api folders are at the root
of your webserver, the default is fine. */
$API_LOCATION = "/API/api.php";

/* The prefix placed at the start of the encrypted part
of each API request. This prefix helps identify that
the request is coming from a panel or Wraith client
and should match the setting in the API database else
requests will be rejected. */
$API_REQUEST_PREFIX = "W_";

/* These are the locations of each page that can be
displayed on the panel. The names of pre-existing
pages are hard-coded in other places and should not
be changed, but other routes can be added and paths
can be edited if you are changing the folder structure
of the panel. Paths are relative to the panel.php
file. */
$PAGE_ROUTES = [
    // Normal pages
    "info" => "./pages/info.php",
    "manager" => "./pages/manager.php",
    "settings" => "./pages/settings.php",
    "usersettings" => "./pages/usersettings.php",
    // Error pages
    "e404" => "./pages/error/404.php",
];

/*
ADDITIONAL CODE

Insert below any additional PHP code you would like
to run every time the panel (not the login page) is
loaded in the browser. This is useful for logging
successful logins for example as this is not done by
default.
*/
