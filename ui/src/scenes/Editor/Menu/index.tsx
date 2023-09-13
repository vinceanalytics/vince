

import { Box, Button, LabelGroup, Label, Spinner, Token } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import {
    TriangleRightIcon, XIcon,
    ZapIcon,
} from "@primer/octicons-react";
import { useQuery, useSites } from "../../../providers";

export const Menu = () => {
    const { selectedSite } = useSites()
    return (
        <Box
            display={"grid"}
            gridTemplateColumns={"1fr auto"}
            alignItems={"center"}
            borderBottomWidth={1}
            borderBottomStyle={"solid"}
            borderBottomColor={"border.default"}
            paddingX={1}
        >
            <Box
                display={"grid"}
                gridTemplateColumns={"auto 1fr"}
                alignItems={"center"}
                sx={{ gap: "1" }}
            >
                <Run />
                <Box>
                    <Token text={selectedSite} />
                </Box>
            </Box>
            <Box>
                <Button
                    variant="primary"
                    leadingIcon={ZapIcon}
                    sx={{ mr: 1 }}
                >Save</Button>
            </Box>
        </Box>
    )
}


const Run = () => {
    const { running, toggleRun } = useQuery()
    return (
        <>
            {running && <Spinner />}
            <Button
                leadingIcon={running ? XIcon : TriangleRightIcon}
                onClick={toggleRun}
                variant={running ? "danger" : "outline"}
            >
                {running ? "Cancel" : "Run"}
            </Button>
        </>
    )
}