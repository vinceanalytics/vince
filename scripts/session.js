

// A user comes on / page, stays there for 10ms and leaves
const case00 = new Session();
case00.wait(10).send();

// A user comes to / page, moves to /careers and leaves
const case01 = new Session();
case01.send().
    with("path", "/careers").send();