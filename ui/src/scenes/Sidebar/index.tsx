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
const Sidebar = () => {
    return (
        <Box
            sx={{
                display: "flex",
                position: "absolute",
                left: "0",
                top: "0",
                paddingLeft: "1rem",
                height: "calc(100% - 4rem)",
                flex: " 0 0 4.5rem",
                flexDirection: "column",
                zIndex: 20001,
                borderRightWidth: 1,
                borderStyle: 'solid',
                borderBottom: "none",
                borderTop: "none",
                borderColor: 'border.default',
            }}>
            <Logo>
                <a href="https://vinceanalytics.github.io" rel="noreferrer" target="_blank">
                    <img alt="VinceAnalytics Logo" height="26" src="/logo.svg" />
                </a>
            </Logo>
            <NavList >
                <Tooltip aria-label="Console" direction="e">
                    <NavList.Item aria-current="page">
                        <NavList.LeadingVisual>
                            <HomeIcon />
                        </NavList.LeadingVisual>
                    </NavList.Item>
                </Tooltip>
                <NavList.Divider sx={{ marginY: "1rem" }} />
                <Tooltip aria-label="Settings" direction="e">
                    <NavList.Item >
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