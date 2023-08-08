import { Box, Text, Link } from '@primer/react'

import { MarkGithubIcon } from "@primer/octicons-react";


const Footer = () => {
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
                <Link
                    href='https://github.com/vinceanalytics/vince'
                    target='_blank'
                    rel='noreferrer'
                >
                    <MarkGithubIcon size={"small"} />
                </Link>
            </Box>
        </Box>
    )
}

export default Footer