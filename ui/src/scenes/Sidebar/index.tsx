import React, { useCallback, useState } from "react"
import { Box, NavList, Portal, Tooltip } from "@primer/react";
import { HomeIcon, GearIcon } from "@primer/octicons-react";



type Tab = "console" | "settings"

export type SideBarProps = {
    onPanelChange: (pane: string) => void,
}
export const Sidebar = ({ onPanelChange }: SideBarProps) => {
    const [selected, setSelected] = useState<Tab>("console")
    const handleConsoleClick = useCallback(() => {
        setSelected("console")
        onPanelChange("console")
    }, [onPanelChange, setSelected])

    const handleSettingsClick = useCallback(() => {
        setSelected("settings")
        onPanelChange("settings")
    }, [onPanelChange, setSelected])

    return (
        <Portal containerName="sidebar"
        >
            <Box
                borderRightWidth={1}
                borderRightStyle={"solid"}
                borderColor={"border.default"}
                height={"100vh"}
            >
                <Box
                    display={"flex"}
                    flex={1}
                    alignItems={"center"}
                    justifyContent={"center"}
                >
                    <a href="https://vinceanalytics.github.io" rel="noreferrer" target="_blank">
                        <img alt="VinceAnalytics Logo" height="32" src="/logo.svg" />
                    </a>
                </Box>
                <NavList >
                    <Tooltip aria-label="Console" direction="e">
                        <NavList.Item
                            aria-current={selected === "console"}
                            onClick={handleConsoleClick}
                        >
                            <NavList.LeadingVisual>
                                <HomeIcon />
                            </NavList.LeadingVisual>
                        </NavList.Item>
                    </Tooltip>
                    <NavList.Divider sx={{ marginY: "1rem" }} />
                    <Tooltip aria-label="Settings" direction="e">
                        <NavList.Item
                            aria-current={selected === "settings"}
                            onClick={handleSettingsClick}
                        >
                            <NavList.LeadingVisual>
                                <GearIcon />
                            </NavList.LeadingVisual>
                        </NavList.Item>
                    </Tooltip>
                </NavList>
            </Box>
        </Portal>
    )
}

