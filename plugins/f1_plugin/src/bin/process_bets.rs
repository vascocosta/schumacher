use f1_plugin::consts;
use f1_plugin::entities::{Bet, EntityManager, RaceResult, User};
use f1_plugin::utils;

fn update_points(users: &Vec<User>, nick: &String, score: u32) -> Vec<User> {
    let mut new_users = Vec::new();

    for user in users {
        if user.nick.to_lowercase() == nick.to_lowercase() {
            new_users.push(User {
                nick: user.nick.to_lowercase(),
                time_zone: user.time_zone.clone(),
                points: user.points + score,
                notifications: user.notifications.clone(),
            });
        } else {
            new_users.push(User {
                nick: user.nick.to_lowercase(),
                time_zone: user.time_zone.clone(),
                points: user.points,
                notifications: user.notifications.clone(),
            });
        }
    }

    new_users
}

fn main() {
    let nick = match utils::parse_args() {
        Ok(nick) => nick,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    if nick.to_lowercase() != "gluon" {
        println!("Only the bot admin can use this command.");

        return;
    }

    let manager = EntityManager::new(consts::PATH);

    let mut race_results = match manager.from_csv::<RaceResult>("race_results") {
        Ok(race_results) => race_results,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    if race_results[0].is_processed() {
        println!("Bets have already been processed for the current event.");

        return;
    }

    let mut bets = match manager.from_csv::<Bet>("bets") {
        Ok(bets) => bets,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    /* Revert back to mutating users, if we fail our mission!
    let mut users = match manager.from_csv::<User>("users") {
        Ok(users) => users,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };
    */

    // This is the main loop where we go through each bet placed by the user and process it.
    // If the race on the bet matches the race on the race results, we calculate its score.
    for bet in bets.iter_mut() {
        
        // Added this here desperately to see if unmut works. REMOVE LATER!
        let users = match manager.from_csv::<User>("users") {
            Ok(users) => users,
            Err(error) => {
                println!("{}", error);
    
                return;
            }
        };

        let mut score = 0;
        if bet.event.to_lowercase() == race_results[0].event.to_lowercase() {
            // If the first driver is on the podium, we have two different possibilities.
            // If the first driver is the first on the results, we score 10.
            // If the first driver is not the first on the results, we score 5.
            if [
                &race_results[0].first,
                &race_results[0].second,
                &race_results[0].third,
            ]
            .contains(&&bet.first)
            {
                // bet.points = 13243245; // (*bet).points = 13243245;
                if bet.first.to_lowercase() == race_results[0].first {
                    score += 10
                } else {
                    score += 5
                }
            }
            // If the second driver is on the podium, we have two different possibilities.
            // If the second driver is the second on the results, we score 10.
            // If the second driver is not the second on the results, we score 5.
            if [
                &race_results[0].first,
                &race_results[0].second,
                &race_results[0].third,
            ]
            .contains(&&bet.second)
            {
                if bet.second.to_lowercase() == race_results[0].second {
                    score += 10
                } else {
                    score += 5
                }
            }
            // If the third driver is on the podium, we have two different possibilities.
            // If the third driver is the second on the results, we score 10.
            // If the third driver is not the second on the results, we score 5.
            if [
                &race_results[0].first,
                &race_results[0].second,
                &race_results[0].third,
            ]
            .contains(&&bet.third)
            {
                if bet.third.to_lowercase() == race_results[0].third {
                    score += 10
                } else {
                    score += 5
                }
            }

            bet.points = score;

            /* Let's try not to mutate the users vector, by creating a function that returns a new one instead!
            for user in users.iter_mut() {
                if user.nick.to_lowercase() == bet.nick.to_lowercase() {
                    user.points += score;
                }
            }
            */

            let new_users = update_points(&users, &bet.nick, score);

            // This was added for unmut. Remove later!
            match manager.to_csv::<User>("users", new_users) {
                Ok(()) => (),
                Err(_) => {
                    println!("Error storing user points.");
        
                    return;
                }
            }
        }
    }

    /*
    for i in 0..bets.len() {
        if bets[i].event.to_lowercase() == race_results[0].event.to_lowercase() {
            if [&race_results[0].first, &race_results[0].second, &race_results[0].third].contains(&&bets[i].first) {
                bets[i].points = 1000;
            }
        }
    }
    */

    /* This is how it was working before umut. Return to this later!
    match manager.to_csv::<User>("users", users) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing user points.");

            return;
        }
    }
    */

    match manager.to_csv::<Bet>("bets", bets) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing bet points.");

            return;
        }
    }

    race_results[0].processed = format!("{}", race_results[0].event);

    match manager.to_csv::<RaceResult>("race_results", race_results) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing last processed bet.");

            return;
        }
    }

    println!("Bets successfully processed.");
}
