import React, { useCallback, useEffect, useState } from "react"
import { Box, NavList } from "@primer/react";
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
                paddingX: "1rem",
                height: "calc(100% - 4rem)",
                flex: " 0 0 4.5rem",
                flexDirection: "column",
                zIndex: 20001,
                width: "130px",
                borderRightWidth: 1,
                borderStyle: 'solid',
                borderColor: 'border.default',
            }}>
            <Logo>
                <a href="https://vinceanalytics.github.io" rel="noreferrer" target="_blank">
                    <img alt="VinceAnalytics Logo" height="26" src="/logo.svg" />
                </a>
            </Logo>
            <NavList >
                <NavList.Item aria-current="page">
                    <NavList.LeadingVisual>
                        <HomeIcon />
                    </NavList.LeadingVisual>
                    Console
                </NavList.Item>
                <NavList.Divider />
                <NavList.Item>
                    <NavList.LeadingVisual>
                        <GearIcon />
                    </NavList.LeadingVisual>
                    Settings
                </NavList.Item>
            </NavList>
        </Box>
    )
}


export default Sidebar