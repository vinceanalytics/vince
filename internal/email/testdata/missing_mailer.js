try {
    __sendMail__(new __Email__());
} catch (error) {
    throw (error.message == "Mailer not configured");
}

