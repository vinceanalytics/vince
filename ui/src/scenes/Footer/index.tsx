import { Box, Text, Link, Label } from '@primer/react'
import { useVince } from "../../providers";
import { MarkGithubIcon } from "@primer/octicons-react";
import { useEffect, useState } from 'react';

const Footer = () => {
    const [version, setVersion] = useState<string>()
    const { vince } = useVince()
    useEffect(() => {
        vince?.version({}).then((result) => {
            setVersion(result.response.version)
        }).catch((e) => { console.log(e) })
    }, [vince, setVersion])
    return (
        <Box id="footer"
            sx={{
                display: "flex",
                position: "absolute",
                height: "4rem",
                bottom: "0",
                left: "0",
                right: "0",
                paddingLeft: "45px",
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

export default Footer