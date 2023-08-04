import React, { useCallback, useEffect, useState } from "react"
import { Box, NavList, Tooltip } from "@primer/react";
import styled from "styled-components"
import { HomeIcon, GearIcon } from "@primer/octicons-react";


const Logo = styled.div`
  position: relative;
  display: flex;
  flex: 0 0 4rem;
  z-index: 1;

  a {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
  }
`

type Tab = "console" | "settings"

const Sidebar = () => {
    const [selected, setSelected] = useState<Tab>("console")
    const handleConsoleClick = useCallback(() => {
        setSelected("console")
    }, [])

    const handleSettingsClick = useCallback(() => {
        setSelected("settings")
    }, [])

    return (
        <Box
            sx={{
                display: "flex",
                position: "absolute",
                left: "0",
                top: "0",
                width: "56px",
                paddingLeft: "1rem",
                height: "calc(100% - 4rem)",
                flex: " 0 0 4.5rem",
                flexDirection: "column",
                zIndex: 20001,
                borderRightWidth: 1,
                borderRightStyle: 'solid',
                borderColor: 'border.default',
            }}>
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
        </Box>
    )
}


export default Sidebar