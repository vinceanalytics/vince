import React, { useCallback, useState } from "react"
import styled from "styled-components"

import { VinceProvider, EditorProvider } from "../../providers";
import Footer from "../Footer";
import { Sidebar } from "../Sidebar"
import Editor from "../Editor"
import { Box, Portal, registerPortalRoot } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import { useLocalStorage, StoreKey, SettingsType } from "../../providers/LocalStorageProvider"
import { Splitter, } from "../../components"
import Sites from "../Sites";
import { Login } from "../Login";


const Top = styled.div`
  position: relative;
  overflow: hidden;
`

registerPortalRoot(document.getElementById("console")!, "console")
registerPortalRoot(document.getElementById("settings")!, "settings")
registerPortalRoot(document.getElementById("login")!, "login")

const Layout = () => {
    const [activePane, setActivePane] = useState<string>("console")

    const { authPayload, editorSplitterBasis, resultsSplitterBasis, updateSettings } =
        useLocalStorage()
    const handleEditorSplitterChange = useCallback((value: SettingsType) => {
        updateSettings(StoreKey.EDITOR_SPLITTER_BASIS, value)
    }, [])

    const handleResultsSplitterChange = useCallback((value: SettingsType) => {
        updateSettings(StoreKey.RESULTS_SPLITTER_BASIS, value)
    }, [])

    const paneChange = (pane: string) => {
        setActivePane(pane)
    }

    return (
        <VinceProvider>
            {authPayload === "" && <Portal containerName="login"><Login /></Portal>}
            {authPayload !== "" &&
                <>
                    <Sidebar onPanelChange={paneChange} />
                    <Footer />
                    <Portal containerName="console">
                        <Box
                            display={activePane === "console" ? "flex" : "none"}
                            flex={1}
                            flexDirection={"column"}
                            maxHeight={"100%"}
                        >
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
                        </Box>
                    </Portal>
                    <Portal containerName="settings">
                        <Box
                            display={activePane === "settings" ? "flex" : "none"}
                            flex={1}
                            flexDirection={"column"}
                            maxHeight={"100%"}
                        >
                            <PageHeader>
                                <PageHeader.TitleArea>
                                    <PageHeader.Title>Settings</PageHeader.Title>
                                </PageHeader.TitleArea>
                            </PageHeader>
                        </Box>
                    </Portal>
                </>}
        </VinceProvider >
    )
}
export default Layout