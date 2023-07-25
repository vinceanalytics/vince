
const case00 = new Session(true);

case00.with("path", "/home").send()
case00.with("path", "/about").send()
case00.wait(10).with("path", "/career").send()

console.log(case00.pretty());
