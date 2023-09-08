import { Box, Text, Link, Label, Portal } from '@primer/react'
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
        <Portal containerName="footer">
            <Box
                display={'grid'}
                gridTemplateColumns={"auto auto"}
            >
                <Box>
                    <Text>
                        Copyright &copy; {new Date().getFullYear()} Vince Analytics
                    </Text>
                </Box>
                <Box
                    display={'grid'}
                    gridTemplateColumns={"auto auto"}
                >
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
        </Portal>
    )
}

export default Footer