

import { Box, Button, LabelGroup, Label, Spinner } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import {
    TriangleRightIcon, XIcon,
    ZapIcon,
} from "@primer/octicons-react";
import { useQuery } from "../../../providers";

export const Menu = () => {
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
            <Box>
                <Run />
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