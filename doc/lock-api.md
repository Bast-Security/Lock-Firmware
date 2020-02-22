
Communication Protocol Between Controller & Lock
================================================

Communication between the lock and the controller will be done through MQTT protocol.

The controller will have an MQTT broker running on it, and the locks and controller will
communicate with one another by publishing and subscribing messages to/from the broker.

Each device is given an MQTT address based on its name.
The hierarchical *root* of the MQTT path is the name 'bast'
This is followed by the name of the controller, then the name of the lock.
Commands append to this path with a verb.

For instance, if the controller is named 'csulb-bast' and the lock is named 'ecs-entrance',
the two devices will communicate over the topic '/bast/csulb-bast/ecs-entrance'.
A command to send a keypad entry would be to '/bast/csulb-bast/ecs-entrance/keypad'.

From-Lock
---------

/keypad - Publish here to send a sequence of keypad inputs to the controller.

/card - Publish here to send a card swipe input to the controller.

/status - Publish here to update the status of the lock.
A value of 1 indicates an alarm status.

/power - Publish here to update the power status of the lock.
This is an integer wich represents the percentage of power left in the battery from 0-100.

To-Lock
-------

/state - Publish here to command the lock to lock.
A value of 0 unlocks the lock.
Any other value will lock the lock.

/channel - Publish here to tell the lock to switch to a different RF channel

/network-id - Publish here to tell the lock to switch to a different network id.

Communication Between Controller and Android App
================================================

Communication between the controller and the android app is facilitated through a rest API.

Endpoints
---------

/challenge - Get an auth challenge (needed to authenticate)

/auth - Respond to the auth challenge by signing it with private key. Returns JWT.
    Response = Signed challenge.

/addUser - Add a user to the system
    name = The name of the user to add
    email = The email address of the new user
    pin = The pin number of the user
    cardno = The card number of the user
    [roles] = names or ids of roles to give this user

/listUsers - Get a list of all the users in JSON format

/addRole - Add a role to the system
    name = The name of the role to add
    [users] = id's of users to add to this role

/listRoles - List all the roles in the system

/listLocks - List all the locks in the system

/listDoors - List all the doors in the system

/nameSystem - Rename the system

/getName - Get the name of this system

/openDoor - Unlock a door
    door = The name of the door to unlock

/closeDoor - Lock a door
    door = The name of the door to lock

/route - The a list of a users activities
    user = id of the user to check

/interrogate - Get a history of users who used a door.
    door = name of the door to interrogate
