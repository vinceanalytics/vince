import { AlertIcon, CodeIcon, TableIcon } from "@primer/octicons-react"
import { BaseStyles, Box, Breadcrumbs, CircleOcticon, NavList, Octicon, Text } from "@primer/react"
import { useEffect, useState } from "react"
import { Item } from "./types"
import { SitesSettingsContext } from "./sites"



export type SettingsProps = {
    activePane: string
}

export const Settings = ({ activePane }: SettingsProps) => {
    const [item, setItem] = useState<Item>("sites")
    const [context, setContext] = useState<string>("")
    return (
        <Box
            display={activePane === "settings" ? "grid" : "none"}
            height={"100%"}
            width={"100%"}
        >
            <Box
                display={"grid"}
                height={"100%"}
                gridTemplateRows={"auto 1fr"}
            >
                <Box id="setting-header"
                    padding={2}
                    borderBottom={1}
                    borderBottomColor={"border.default"}
                    borderBottomStyle={"solid"}
                >
                    <Text>Settings</Text>
                </Box>
                <Box id="settings-body"
                    display={"grid"}
                    gridTemplateColumns={"auto auto 1fr"}
                    sx={{ gap: "2" }}
                    height={"100vh"}
                >
                    <Box id="settings-menu"
                        padding={2}
                        borderRightWidth={1}
                        borderRightStyle={"solid"}
                        borderColor={"border.default"}
                    >
                        <NavList>
                            <NavList.Item
                                aria-current={item === "sites"}
                                onClick={() => setItem("sites")}
                            >
                                Sites
                            </NavList.Item>
                            <NavList.Item
                                aria-current={item === "snippets"}
                                onClick={() => setItem("snippets")}
                            >
                                Snippets
                            </NavList.Item>
                            <NavList.Item
                                aria-current={item === "alerts"}
                                onClick={() => setItem("alerts")}
                            >
                                Alerts
                            </NavList.Item>
                        </NavList>
                    </Box>
                    <Box id="settings-menu-context"
                        padding={2}
                        borderRightWidth={1}
                        borderRightStyle={"solid"}
                        borderColor={"border.default"}
                    >
                        <SitesSettingsContext item={item} />
                    </Box>
                    <Box id="settings-menu-view"></Box>
                </Box>
            </Box>
        </Box>
    )
}