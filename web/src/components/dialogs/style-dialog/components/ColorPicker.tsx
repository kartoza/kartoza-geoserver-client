import {
  Box,
  FormControl,
  FormLabel,
  HStack,
  Input,
  useColorModeValue,
} from '@chakra-ui/react'

interface ColorPickerProps {
  value: string
  onChange: (color: string) => void
  label: string
}

export function ColorPicker({ value, onChange, label }: ColorPickerProps) {
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  return (
    <FormControl>
      <FormLabel fontSize="sm">{label}</FormLabel>
      <HStack>
        <Box position="relative">
          <Box
            w="40px"
            h="40px"
            borderRadius="full"
            bg={value}
            border="2px solid"
            borderColor={borderColor}
            cursor="pointer"
            _hover={{ borderColor: 'kartoza.400' }}
            transition="border-color 0.2s"
          />
          <Input
            type="color"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            position="absolute"
            top={0}
            left={0}
            w="40px"
            h="40px"
            opacity={0}
            cursor="pointer"
          />
        </Box>
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          size="sm"
          w="100px"
          fontFamily="mono"
          borderColor={borderColor}
        />
      </HStack>
    </FormControl>
  )
}
