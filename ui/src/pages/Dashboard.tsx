import { ChevronDownIcon } from "@primer/octicons-react"
import { PageLayout, Box, Heading, Button, Octicon, ActionMenu, ActionList, Text, TokenProps, TextInputWithTokens, Select, FormControl, TextInput } from "@primer/react"
import { useCallback, useState } from "react"
import { Dialog } from '@primer/react/drafts'
import { Footer, CurrentVisitors, SitesSelector } from "../components";
import { useVince } from "../providers";


export const Dashboard = () => {
    const { sites, active, selectSite } = useVince()
    const [tokens, setTokens] = useState<TokenProps[]>([])

    const onTokenRemove = (tokenId: string | number) => {
        setTokens(tokens.filter(token => token.id !== tokenId))
    }

    const onTokenAdd = (tokenId: string) => {
        setTokens([{ id: tokens.length, text: tokenId }, ...tokens])
    }

    return (
        <PageLayout>
            <PageLayout.Header>
                <Box py={2} sx={{
                    display: "grid",
                    alignItems: "center",
                    gridTemplateColumns: "auto auto 1fr auto auto",
                    gap: 1,
                }}>
                    <SitesSelector selectSite={selectSite} active={active} sites={sites} />
                    <CurrentVisitors />
                    <div></div>
                    <Filter tokens={tokens} onAdd={onTokenAdd} onRemove={onTokenRemove} />
                    <DatePicker />
                </Box>
            </PageLayout.Header>
            <PageLayout.Content>
                <Box m={4}>
                    <Heading as="h2" sx={{ mb: 2 }}>
                        Hello, world!
                    </Heading>
                    <p>This will get Primer text styles.</p>
                </Box>
            </PageLayout.Content>
            <PageLayout.Footer>
                <Footer />
            </PageLayout.Footer>
        </PageLayout>
    )
}





type FilterProps = {
    tokens: TokenProps[]
    onRemove: (tokenId: string | number) => void
    onAdd: (tokenId: string) => void
}

const Filter = ({ tokens, onRemove, onAdd }: FilterProps) => {
    const [isOpen, setOpen] = useState<boolean>(false)
    const openDialog = useCallback(() => setOpen(true), [setOpen])
    const closeDialog = useCallback(() => setOpen(false), [setOpen])

    return (
        <Box>
            {isOpen && (
                <Dialog
                    title="Add filter"
                    onClose={closeDialog}
                >
                    <FilterItems onAdd={(filter) => {
                        onAdd(filter)
                        closeDialog()
                    }} />
                </Dialog>
            )}
            <TextInputWithTokens placeholder="filters" tokens={tokens} onTokenRemove={onRemove}
                onClick={openDialog}
            />
        </Box>
    )
}





const properties = [
    "event",
    "page",
    "entry_page",
    "exit_page",
    "source",
    "referrer",
    "utm_source",
    "utm_medium",
    "utm_campaign",
    "utm_content",
    "utm_term",
    "device",
    "browser",
    "browser_version",
    "os",
    "os_version",
    "country",
    "region",
    "domain",
    "city"
]

interface Op {
    name: string
    value: string
}

const ops: Op[] = [
    { name: "equal", value: "==" },
    { name: "not_equal", value: "!=" },
    { name: "regex_equal", value: "~=" },
    { name: "regex_not_equal", value: "!~" }
]



type FilterItemsProps = {
    onAdd: (filter: string) => void
}

const FilterItems = ({ onAdd }: FilterItemsProps) => {
    const [property, setProperty] = useState<string>()
    const [op, setOp] = useState<string>()
    const [value, steValue] = useState<string>()
    const submit = useCallback(() => {
        if (property && op && value) {
            onAdd(property + op + value)
        }
    }, [property, op, value, onAdd])
    return (
        <Box display={"flex"}>
            <Select placeholder="Select Property" onChange={(e) => {
                setProperty(e.target.value)
            }}>
                {properties.map((value) =>
                    <Select.Option key={value} value={value}>{value}</Select.Option>
                )}
            </Select>
            <Select placeholder="Select Operation" onChange={(e) => {
                setOp(e.target.value)
            }}>
                {ops.map((value) =>
                    <Select.Option key={value.value} value={value.value}>{value.name}</Select.Option>
                )}
            </Select>
            <TextInput placeholder="Value" onChange={(e) => {
                steValue(e.target.value)
            }} required>
            </TextInput>
            <Button variant="primary" sx={{ px: 5, mx: 2 }} onClick={submit} >Add</Button>
        </Box>
    )
}


const DatePicker = () => {
    return (
        <Box>
            <ActionMenu>
                <ActionMenu.Button><Text sx={{ fontWeight: "bold" }}>Today</Text></ActionMenu.Button>
                <ActionMenu.Overlay side="outside-left">
                    <ActionList>
                        <ActionList.Item  ><Text sx={{ fontWeight: "bold" }}>Last 15 days</Text></ActionList.Item>
                    </ActionList>
                </ActionMenu.Overlay>
            </ActionMenu>
        </Box>
    )
}