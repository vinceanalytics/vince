import { useState, useCallback, ReactNode } from "react";
import { Text, TreeView, Box, themeGet, IconButton, Octicon } from "@primer/react";
import { PlusIcon, DatabaseIcon } from "@primer/octicons-react";
import { Dialog } from '@primer/react/drafts'

import styled, { css } from "styled-components"

export const PaneMenu = styled.div`
  position: relative;
  display: flex;
  height: 3rem;
  padding: 0 1rem;
  align-items: center;
  background: ${themeGet("canvas.default")};
  border-top: 1px solid transparent;
  z-index: 5;
`

const loadingStyles = css`
  display: flex;
  justify-content: center;
`

export const PaneWrapper = styled.div`
  display: flex;
  flex-direction: column;
  flex: 1;
`

const Wrapper = styled(PaneWrapper)`
  overflow-x: auto;
  height: 100%;
`

const Menu = styled(PaneMenu)`
  justify-content: space-between;
`


const Header = styled(Text)`
  display: flex;
  align-items: center;
`

type Props = Readonly<{
    children: ReactNode
    className?: string
}>

export const PaneContent = styled.div<Props>`
    display: flex;
    flex-direction: column;
    flex: 1;
    background: ${themeGet("canvas.default")};
    overflow: auto;
  `

const Content = styled(PaneContent) <{
    _loading: boolean
}>`
    display: block;
    overflow: auto;
    ${({ _loading }) => _loading && loadingStyles};
  `




const Sites = () => {
    return (
        <Wrapper>
            <Box sx={{ borderBottomWidth: 1, borderBottomStyle: 'solid', borderColor: 'border.default' }}>
                < Menu >
                    <Header>
                        <Octicon icon={DatabaseIcon} sx={{ marginRight: "1rem" }} />
                        Sites
                    </Header>
                    <Box display={"flex"}>
                        <IconButton size="small" aria-label="new site" icon={PlusIcon} />
                    </Box>
                </Menu>
            </Box >
        </Wrapper >
    )
}

export default Sites;