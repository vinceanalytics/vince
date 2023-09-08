import React, { useCallback, useEffect, useState } from "react"
import { Box, NavList, Portal, Tooltip } from "@primer/react";
import styled from "styled-components"
import { HomeIcon, GearIcon } from "@primer/octicons-react";


const Logo = styled.div`
  position: relative;
  display: flex;
  flex: 0 0 3rem;
  z-index: 1;

  a {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
  }
`

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
            <Logo>
                <a href="https://vinceanalytics.github.io" rel="noreferrer" target="_blank">
                    <img alt="VinceAnalytics Logo" height="26" src="/logo.svg" />
                </a>
            </Logo>
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
        </Portal>
    )
}

