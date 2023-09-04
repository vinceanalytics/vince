
import { Box, Button, Portal, TextInput, registerPortalRoot } from "@primer/react";
import { useCallback, useState } from "react";
import { login } from "../../vince";
import { useLocalStorage, StoreKey } from "../../providers/LocalStorageProvider";

const Login = () => {
    const { updateSettings } = useLocalStorage()
    const [userName, setUserName] = useState<string>("")
    const [password, setPassWord] = useState<string>("")
    const [loading, setLoading] = useState<boolean>(false)
    const submit = useCallback(() => {
        setLoading(true)
        login(userName, password).then((result) => {
            updateSettings(StoreKey.AUTH_PAYLOAD, result.response.auth?.token!)
        }).catch((e) => {
            console.log(e)
        })
    }, [userName, password, setLoading])
    return (
        <Box
            display={"flex"}
            width={"100%"}
            height={"100vh"}
            justifyContent={"center"}
            alignItems={"center"}
        >
            <form>
                <Box display={"flex"}>
                    <Box display={"flex"} flex={1} alignItems={"center"} justifyContent={"center"} mr={1}>
                        <a href="https://vinceanalytics.github.io" rel="noreferrer" target="_blank">
                            <img alt="VinceAnalytics Logo" height="26" src="/logo.svg" />
                        </a>
                    </Box>
                    <TextInput
                        name="username"
                        aria-label="vince root account name"
                        required
                        onChange={(e) => setUserName(e.currentTarget.value)}
                        loading={loading ? true : undefined}
                        autoComplete="username"
                        monospace placeholder="username" sx={{ marginRight: 1 }} />
                    <TextInput
                        name="password"
                        aria-label="vince root account password"
                        required
                        onChange={(e) => setPassWord(e.currentTarget.value)}
                        loading={loading ? true : undefined}
                        autoComplete="current-password"
                        monospace placeholder="password" type="password" sx={{ marginRight: 1 }} />
                    <Button variant="primary" onClick={submit}>Login</Button>
                </Box>
            </form>
        </Box>
    )
}

registerPortalRoot(document.getElementById("login")!, "login")

export const Auth = (props: { children?: React.ReactNode; }) => {
    const { children } = props;
    const { authPayload } = useLocalStorage()
    return (
        <>
            {authPayload == "" && <Portal containerName="login"><Login /></Portal>}
            {authPayload !== "" && children}
        </>
    )
}