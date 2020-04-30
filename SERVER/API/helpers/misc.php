<?php

// Get a list of all the available command modules
function get_cmd_modules() {
	$command_module_paths = glob('storage/command_modules/*.wgo');
	$command_modules = [];
	foreach ($command_module_paths as $path) {
		// Get the contents of the file (script)
		$command_contents = file_get_contents($path);
		$command_modules[basename($path, ".wgo")] = [$path, $command_contents];
	}
	return $command_modules;
}

// Get IP of client
function get_client_ip() {
	$ipaddress = 'UNKNOWN';
	$keys = array('HTTP_CLIENT_IP','HTTP_X_FORWARDED_FOR','HTTP_X_FORWARDED','HTTP_FORWARDED_FOR','HTTP_FORWARDED','REMOTE_ADDR');
	foreach($keys as $k)
	{
		if (isset($_SERVER[$k]) && !empty($_SERVER[$k]) && filter_var($_SERVER[$k], FILTER_VALIDATE_IP))
		{
			$ipaddress = $_SERVER[$k];
			break;
		}
	}
	return $ipaddress;
}
