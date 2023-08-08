import React, { useCallback, useEffect } from "react"
import { createPortal } from "react-dom"
import styled from "styled-components"

import { VinceProvider, EditorProvider } from "../../providers";
import Footer from "../Footer";
import Sidebar from "../Sidebar"
import Editor from "../Editor"
import { Box } from "@primer/react";
import { useLocalStorage, StoreKey, SettingsType } from "../../providers/LocalStorageProvider"
import { Splitter, } from "../../components"
import Sites from "../Sites";
const Console = styled.div`
  display: flex;
  flex-direction: column;
  flex: 1;
  max-height: 100%;
`

const Top = styled.div`
  position: relative;
  overflow: hidden;
`

const Layout = () => {
    const consoleNode = document.getElementById("console")
    const { editorSplitterBasis, resultsSplitterBasis, updateSettings } =
        useLocalStorage()
    const handleEditorSplitterChange = useCallback((value: SettingsType) => {
        updateSettings(StoreKey.EDITOR_SPLITTER_BASIS, value)
    }, [])

    const handleResultsSplitterChange = useCallback((value: SettingsType) => {
        updateSettings(StoreKey.RESULTS_SPLITTER_BASIS, value)
    }, [])
    return (
        <VinceProvider>
            <Sidebar />
            <Footer />
            {consoleNode &&
                createPortal(
                    <Console>
                        <EditorProvider>
                            <Splitter
                                direction="vertical"
                                fallback={editorSplitterBasis}
                                min={100}
                                onChange={handleEditorSplitterChange}
                            >
                                <Top>
                                    <Splitter
                                        direction="horizontal"
                                        fallback={resultsSplitterBasis}
                                        max={500}
                                        onChange={handleResultsSplitterChange}
                                    >
                                        <Box >
                                            <Sites />
                                        </Box>
                                        <Editor />
                                    </Splitter>
                                </Top>
                                <Box />
                            </Splitter>
                        </EditorProvider>
                    </Console>,
                    consoleNode,
                )
            }
        </VinceProvider>
    )
}
export default Layout