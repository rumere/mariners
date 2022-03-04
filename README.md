# mariners

App to track Mariners Point golf league data

## TODO List

* update index.html to use `.Title` for the main frame
  * skip handlers for rendered templates...
  * set the title to the path name, and use showsection...
  * this will fix the "back" button, i think.
  * also, will allow for direct links and keep the refresh button from going back home.
* make authentication middleware
* make inviteonly a real thing :)
* remove payment conbtrols from unpaid events
* check event dates for visibility
* more permisiions checking for secondary roles
* also check permissions on the server side, duh.
* Create a background task for -
  * event cleanup
  * maintain the "user" SNS Topic
