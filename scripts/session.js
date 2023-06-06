const session = new Session();

session.fixture = true;
println(session.send().requests.dump());