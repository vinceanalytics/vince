let e = new __Email__();

e.from.name = "jane";
e.from.address = "jane@example.com";
e.to.name = "doe";
e.to.address = "doe@example.com";
e.subject = "Testing";
e.msg = "Hello, world";
__sendMail__(e);