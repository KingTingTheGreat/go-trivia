# Trivia

A live buzzer to play trivia on your local network

## Setup

-   Install [Go version 1.22 or higher](https://go.dev/doc/install)
-   Install [templ](https://templ.guide/quick-start/installation/) and [air](https://github.com/cosmtrek/air)
-   Create a .env file and add a password; PASSWORD="yourpasswordhere"
-   Run "templ generate" this should create *_templ.go files in the templates directory
-   Run "air" or "go run main.go" to serve the application
-   The application is now accessible on your local network at your device's IP on port 3000!

## How to use

### Players

-   Visit HostIP:3000 while on the same network
-   Enter your name and advance to the next page
-   You will now see a buzzer, clicking it will notify the host

### Host

-   Visit HostIP:3000/host to view the players who have buzzed in sorted in chronological order, as well as the time they buzzed in down to the millisecond
-   This page will also display the players ranked by score, score seniority is used as a tie-breaker when scores are equal
-   Go to HostIP:3000/control to access controls
    -   Enter the password you put in your .env file
    -   Enter a player's name and the number of points you'd like to give them
        -   A negative value will subtract from their score while a positive value will add. Negative scores are possible
        -   A positive value will append the current question number to the questions the player has gotten correct
        -   A zero or negative value will append a negative question number to the questions the player has gotten correct
    -   Clear will clear the current buzzed in list but will not affect anything else
    -   Next will increment the question number and clear the buzzed in list

### Additional Pages

#### Leaderboard

-   Used to only view the players ranked by score

#### Stats

-   View players ranked by score as well as the questions each player has received points for

#### Buzzed-In

-   See the players as they buzz in for the current question
