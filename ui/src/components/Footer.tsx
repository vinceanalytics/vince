import { useEffect, useState } from "react"
import { useVince } from "../providers"
import { Box, Link, Label, Text } from "@primer/react"
import { MarkGithubIcon } from "@primer/octicons-react"


export const Footer = () => {
    const { vince } = useVince()
    const [version, setVersion] = useState<string>("unknown")
    useEffect(() => {
        vince.version().then((value) => {
            setVersion(value.version)
        }).catch(console.log)
    }, [vince])
    return (
        <Box
            sx={{
                display: "flex",
                height: "4rem",
                bottom: "0",
                left: "0",
                right: "0",
                backgroundColor: 'canvas.subtle',
            }}
        >
            <Box
                sx={{
                    display: "flex",
                    paddingLeft: "1rem",
                    alignItems: "center",
                    flex: "1",
                }}
            >
                <Text>
                    Copyright &copy; {new Date().getFullYear()} Vince Analytics
                </Text>
            </Box>
            <Box sx={{
                display: "flex",
                paddingRight: "1rem",
                alignItems: "center",
            }}>
                <Label variant="primary" sx={{
                    marginRight: 1,
                }}>
                    vince: {version}
                </Label>
                <Link
                    href='https://github.com/vinceanalytics/vince'
                    target='_blank'
                    rel='noreferrer'
                >
                    <MarkGithubIcon size={"medium"} />
                </Link>
            </Box>
        </Box>
    )
}