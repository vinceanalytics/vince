import { Box, BoxProps, TextProps, Text, HeadingProps, Heading } from "@primer/react";
import { PropsWithChildren } from "react";


export const Card = ({ children, ...rest }: PropsWithChildren<BoxProps>) => {
    return (
        <Box
            borderWidth={1}
            borderStyle={"solid"}
            borderColor={"border.default"}
            borderRadius={"1rem"}
            {...rest}
        >
            {children}
        </Box>
    )
}

export const CardTitle = ({ children, ...rest }: PropsWithChildren<HeadingProps>) => {
    return (
        <Heading
            as="h3"
            sx={{ fontWeight: "semibold" }}
            {...rest}
        >
            {children}
        </Heading>
    )
}

export const CardHeader = ({ children, ...rest }: PropsWithChildren<BoxProps>) => {
    return (
        <Box
            display={"flex"}
            flexDirection={"column"}
            p={4}
            {...rest}
        >
            {children}
        </Box>
    )
}

export const CardDescription = ({ children, ...rest }: PropsWithChildren<TextProps>) => {
    return (
        <Text
            fontSize={1}
            color={"fg.muted"}
            {...rest}
        >
            {children}
        </Text>
    )
}


export const CardContent = ({ children, ...rest }: PropsWithChildren<BoxProps>) => {
    return (
        <Box
            p={4}
            pt={0}
            {...rest}
        >
            {children}
        </Box>
    )
}

export const CardFooter = ({ children, ...rest }: PropsWithChildren<BoxProps>) => {
    return (
        <Box
            display={"flex"}
            alignItems={"center"}
            p={6}
            pt={0}
            {...rest}
        >
            {children}
        </Box>
    )
}
