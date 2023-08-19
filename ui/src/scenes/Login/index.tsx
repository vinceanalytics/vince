
import { Box, Button, TextInput } from "@primer/react";

export const Login = () => {
    return (
        <Box
            display={"flex"}
            width={"100%"}
            height={"100vh"}
            justifyContent={"center"}
            alignItems={"center"}
        >
            <Box display={"flex"}>
                <Box display={"flex"} flex={1} alignItems={"center"} justifyContent={"center"} mr={1}>
                    <a href="https://vinceanalytics.github.io" rel="noreferrer" target="_blank">
                        <img alt="VinceAnalytics Logo" height="26" src="/logo.svg" />
                    </a>
                </Box>
                <TextInput placeholder="username" sx={{ marginRight: 1 }} />
                <TextInput placeholder="password" type="password" sx={{ marginRight: 1 }} />
                <Button variant="primary">Login</Button>
            </Box>
        </Box>
    )
}