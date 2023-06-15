let e = new __Email__();

e.to.name = "doe";
e.to.address = "doe@example.com";
e.subject = "Testing";
e.msg = "Hello, world";
__sendMail__(e);