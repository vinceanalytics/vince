import styled from "styled-components"
import React, { CSSProperties, forwardRef, Ref } from "react"
import { BoxProps, Box } from "@primer/react";
import Monaco from "./Monaco"
import { Menu } from "./Menu";



export const PaneWrapper = styled.div`
  display: flex;
  flex-direction: column;
  flex: 1;
`

const EditorPaneWrapper = styled(PaneWrapper)`
  overflow: hidden;
`

const Editor = ({
    innerRef,
    ...rest
}: BoxProps & { innerRef: Ref<HTMLDivElement> }) => (
    <Box
        display={"flex"}
        flexDirection={"column"}
        flex={1}
        overflow={"hidden"}
        ref={innerRef} {...rest}
    >
        <Menu />
        <Monaco />
    </Box>
)

const EditorWithRef = (props: BoxProps, ref: Ref<HTMLDivElement>) => (
    <Editor {...props} innerRef={ref} />
)

export default forwardRef(EditorWithRef)