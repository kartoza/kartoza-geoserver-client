import { useState } from 'react'
import {
  Box,
  Button,
  Collapse,
  Divider,
  Flex,
  FormControl,
  FormLabel,
  HStack,
  Icon,
  IconButton,
  Input,
  NumberDecrementStepper,
  NumberIncrementStepper,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  Select,
  Slider,
  SliderFilledTrack,
  SliderThumb,
  SliderTrack,
  Text,
  useColorModeValue,
  VStack,
} from '@chakra-ui/react'
import { FiChevronDown, FiChevronUp, FiMinus } from 'react-icons/fi'
import type { StyleRule } from '../types'
import { POINT_STYLE_PRESETS } from '../constants'
import { ColorPicker } from './ColorPicker'

interface RuleEditorProps {
  rule: StyleRule
  onChange: (rule: StyleRule) => void
  onDelete: () => void
}

export function RuleEditor({ rule, onChange, onDelete }: RuleEditorProps) {
  const bgColor = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const [showMorePresets, setShowMorePresets] = useState(false)

  const updateSymbolizer = (updates: Partial<StyleRule['symbolizer']>) => {
    onChange({
      ...rule,
      symbolizer: { ...rule.symbolizer, ...updates }
    })
  }

  return (
    <Box
      p={4}
      bg={bgColor}
      borderRadius="lg"
      border="1px solid"
      borderColor={borderColor}
    >
      <VStack spacing={4} align="stretch">
        <HStack justify="space-between">
          <FormControl maxW="200px">
            <FormLabel fontSize="sm">Rule Name</FormLabel>
            <Input
              size="sm"
              value={rule.name}
              onChange={(e) => onChange({ ...rule, name: e.target.value })}
            />
          </FormControl>
          <FormControl maxW="150px">
            <FormLabel fontSize="sm">Geometry Type</FormLabel>
            <Select
              size="sm"
              value={rule.symbolizer.type}
              onChange={(e) => updateSymbolizer({ type: e.target.value as 'polygon' | 'line' | 'point' })}
            >
              <option value="polygon">Polygon</option>
              <option value="line">Line</option>
              <option value="point">Point</option>
            </Select>
          </FormControl>
          <IconButton
            aria-label="Delete rule"
            icon={<FiMinus />}
            size="sm"
            colorScheme="red"
            variant="ghost"
            onClick={onDelete}
          />
        </HStack>

        <Divider />

        {/* Fill settings (for polygon and point) */}
        {(rule.symbolizer.type === 'polygon' || rule.symbolizer.type === 'point') && (
          <Box>
            <Text fontWeight="600" fontSize="sm" mb={2}>Fill</Text>
            <HStack spacing={4} wrap="wrap">
              <ColorPicker
                label="Color"
                value={rule.symbolizer.fill || '#3388ff'}
                onChange={(color) => updateSymbolizer({ fill: color })}
              />
              <FormControl maxW="150px">
                <FormLabel fontSize="sm">Opacity</FormLabel>
                <HStack>
                  <Slider
                    value={rule.symbolizer.fillOpacity ?? 1}
                    min={0}
                    max={1}
                    step={0.1}
                    onChange={(val) => updateSymbolizer({ fillOpacity: val })}
                  >
                    <SliderTrack>
                      <SliderFilledTrack bg="kartoza.500" />
                    </SliderTrack>
                    <SliderThumb />
                  </Slider>
                  <Text fontSize="sm" w="40px">{((rule.symbolizer.fillOpacity ?? 1) * 100).toFixed(0)}%</Text>
                </HStack>
              </FormControl>
            </HStack>
          </Box>
        )}

        {/* Stroke settings */}
        <Box>
          <Text fontWeight="600" fontSize="sm" mb={2}>Stroke</Text>
          <HStack spacing={4} wrap="wrap">
            <ColorPicker
              label="Color"
              value={rule.symbolizer.stroke || '#2266cc'}
              onChange={(color) => updateSymbolizer({ stroke: color })}
            />
            <FormControl maxW="100px">
              <FormLabel fontSize="sm">Width</FormLabel>
              <NumberInput
                size="sm"
                value={rule.symbolizer.strokeWidth || 1}
                min={0}
                max={20}
                step={0.5}
                onChange={(_, val) => updateSymbolizer({ strokeWidth: val })}
              >
                <NumberInputField />
                <NumberInputStepper>
                  <NumberIncrementStepper />
                  <NumberDecrementStepper />
                </NumberInputStepper>
              </NumberInput>
            </FormControl>
            {rule.symbolizer.type === 'line' && (
              <FormControl maxW="150px">
                <FormLabel fontSize="sm">Opacity</FormLabel>
                <HStack>
                  <Slider
                    value={rule.symbolizer.strokeOpacity ?? 1}
                    min={0}
                    max={1}
                    step={0.1}
                    onChange={(val) => updateSymbolizer({ strokeOpacity: val })}
                  >
                    <SliderTrack>
                      <SliderFilledTrack bg="kartoza.500" />
                    </SliderTrack>
                    <SliderThumb />
                  </Slider>
                  <Text fontSize="sm" w="40px">{((rule.symbolizer.strokeOpacity ?? 1) * 100).toFixed(0)}%</Text>
                </HStack>
              </FormControl>
            )}
          </HStack>
        </Box>

        {/* Point-specific settings */}
        {rule.symbolizer.type === 'point' && (
          <Box>
            <Text fontWeight="600" fontSize="sm" mb={2}>Point Symbol</Text>

            {/* Style Presets Gallery */}
            <Box mb={4}>
              <Text fontSize="xs" color="gray.500" mb={2}>Quick Presets</Text>
              <Flex flexWrap="wrap" gap={2}>
                {POINT_STYLE_PRESETS.slice(0, 10).map((preset) => (
                  <Box
                    key={preset.name}
                    title={`${preset.name}: ${preset.description}`}
                    cursor="pointer"
                    p={1}
                    borderRadius="md"
                    border="2px solid"
                    borderColor={
                      rule.symbolizer.fill === preset.fill &&
                      rule.symbolizer.pointShape === preset.shape
                        ? 'kartoza.500'
                        : 'transparent'
                    }
                    _hover={{ borderColor: 'kartoza.300', bg: 'gray.50' }}
                    onClick={() => {
                      updateSymbolizer({
                        pointShape: preset.shape as 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x',
                        fill: preset.fill,
                        fillOpacity: preset.fillOpacity,
                        stroke: preset.stroke,
                        strokeWidth: preset.strokeWidth,
                        pointSize: preset.size,
                        haloColor: preset.haloColor,
                        haloRadius: preset.haloRadius,
                        rotation: preset.rotation,
                      })
                    }}
                  >
                    <Box
                      w="28px"
                      h="28px"
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      position="relative"
                    >
                      {/* Halo effect */}
                      {preset.haloColor && (
                        <Box
                          position="absolute"
                          w={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                          h={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                          borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                          bg={preset.haloColor}
                          opacity={0.4}
                          transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                          style={{
                            clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' : undefined,
                          }}
                        />
                      )}
                      {/* Main symbol */}
                      <Box
                        position="relative"
                        w={`${preset.size}px`}
                        h={`${preset.size}px`}
                        borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                        bg={preset.fill}
                        opacity={preset.fillOpacity}
                        border={`${preset.strokeWidth}px solid ${preset.stroke}`}
                        transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                        style={{
                          clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                                   preset.shape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                                   preset.shape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                                   undefined,
                        }}
                      />
                    </Box>
                  </Box>
                ))}
              </Flex>

              {/* Show more presets */}
              <Collapse in={showMorePresets} animateOpacity>
                <Flex flexWrap="wrap" gap={2} mt={2}>
                  {POINT_STYLE_PRESETS.slice(10).map((preset) => (
                    <Box
                      key={preset.name}
                      title={`${preset.name}: ${preset.description}`}
                      cursor="pointer"
                      p={1}
                      borderRadius="md"
                      border="2px solid"
                      borderColor={
                        rule.symbolizer.fill === preset.fill &&
                        rule.symbolizer.pointShape === preset.shape
                          ? 'kartoza.500'
                          : 'transparent'
                      }
                      _hover={{ borderColor: 'kartoza.300', bg: 'gray.50' }}
                      onClick={() => {
                        updateSymbolizer({
                          pointShape: preset.shape as 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x',
                          fill: preset.fill,
                          fillOpacity: preset.fillOpacity,
                          stroke: preset.stroke,
                          strokeWidth: preset.strokeWidth,
                          pointSize: preset.size,
                          haloColor: preset.haloColor,
                          haloRadius: preset.haloRadius,
                          rotation: preset.rotation,
                        })
                      }}
                    >
                      <Box
                        w="28px"
                        h="28px"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                        position="relative"
                      >
                        {preset.haloColor && (
                          <Box
                            position="absolute"
                            w={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                            h={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                            borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                            bg={preset.haloColor}
                            opacity={0.4}
                            transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                            style={{
                              clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' : undefined,
                            }}
                          />
                        )}
                        <Box
                          position="relative"
                          w={`${preset.size}px`}
                          h={`${preset.size}px`}
                          borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                          bg={preset.fill}
                          opacity={preset.fillOpacity}
                          border={`${preset.strokeWidth}px solid ${preset.stroke}`}
                          transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                          style={{
                            clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                                     preset.shape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                                     preset.shape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                                     undefined,
                          }}
                        />
                      </Box>
                    </Box>
                  ))}
                </Flex>
              </Collapse>

              <Button
                size="xs"
                variant="ghost"
                mt={2}
                onClick={() => setShowMorePresets(!showMorePresets)}
                rightIcon={<Icon as={showMorePresets ? FiChevronUp : FiChevronDown} />}
              >
                {showMorePresets ? 'Show Less' : `Show ${POINT_STYLE_PRESETS.length - 10} More`}
              </Button>
            </Box>

            <Divider my={3} />

            {/* Manual controls */}
            <HStack spacing={4} wrap="wrap">
              <FormControl maxW="150px">
                <FormLabel fontSize="sm">Shape</FormLabel>
                <Select
                  size="sm"
                  value={rule.symbolizer.pointShape || 'circle'}
                  onChange={(e) => updateSymbolizer({ pointShape: e.target.value as 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x' })}
                >
                  <option value="circle">Circle</option>
                  <option value="square">Square</option>
                  <option value="triangle">Triangle</option>
                  <option value="star">Star</option>
                  <option value="cross">Cross</option>
                  <option value="x">X</option>
                </Select>
              </FormControl>
              <FormControl maxW="100px">
                <FormLabel fontSize="sm">Size</FormLabel>
                <NumberInput
                  size="sm"
                  value={rule.symbolizer.pointSize || 8}
                  min={1}
                  max={50}
                  onChange={(_, val) => updateSymbolizer({ pointSize: val })}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
              <FormControl maxW="100px">
                <FormLabel fontSize="sm">Rotation</FormLabel>
                <NumberInput
                  size="sm"
                  value={rule.symbolizer.rotation || 0}
                  min={0}
                  max={360}
                  step={15}
                  onChange={(_, val) => updateSymbolizer({ rotation: val })}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </HStack>

            {/* Halo / Glow effect */}
            <Box mt={4}>
              <Text fontWeight="600" fontSize="sm" mb={2}>Halo / Glow Effect</Text>
              <HStack spacing={4} wrap="wrap">
                <ColorPicker
                  label="Halo Color"
                  value={rule.symbolizer.haloColor || '#ffffff'}
                  onChange={(color) => updateSymbolizer({ haloColor: color })}
                />
                <FormControl maxW="120px">
                  <FormLabel fontSize="sm">Halo Radius</FormLabel>
                  <HStack>
                    <Slider
                      value={rule.symbolizer.haloRadius || 0}
                      min={0}
                      max={10}
                      step={1}
                      onChange={(val) => updateSymbolizer({ haloRadius: val })}
                    >
                      <SliderTrack>
                        <SliderFilledTrack bg="kartoza.500" />
                      </SliderTrack>
                      <SliderThumb />
                    </Slider>
                    <Text fontSize="sm" w="30px">{rule.symbolizer.haloRadius || 0}px</Text>
                  </HStack>
                </FormControl>
                {rule.symbolizer.haloRadius && rule.symbolizer.haloRadius > 0 && (
                  <Button
                    size="xs"
                    variant="ghost"
                    colorScheme="red"
                    onClick={() => updateSymbolizer({ haloRadius: 0, haloColor: undefined })}
                  >
                    Remove Halo
                  </Button>
                )}
              </HStack>
            </Box>
          </Box>
        )}

        {/* Preview swatch */}
        <Box>
          <Text fontWeight="600" fontSize="sm" mb={2}>Preview</Text>
          <Box
            w="100px"
            h="60px"
            borderRadius="md"
            border="1px solid"
            borderColor={borderColor}
            display="flex"
            alignItems="center"
            justifyContent="center"
            bg="gray.100"
          >
            {rule.symbolizer.type === 'polygon' && (
              <Box
                w="60px"
                h="40px"
                borderRadius="sm"
                bg={rule.symbolizer.fill}
                opacity={rule.symbolizer.fillOpacity}
                border={`${rule.symbolizer.strokeWidth}px solid ${rule.symbolizer.stroke}`}
              />
            )}
            {rule.symbolizer.type === 'line' && (
              <Box
                w="60px"
                h={`${Math.max(2, rule.symbolizer.strokeWidth || 2)}px`}
                bg={rule.symbolizer.stroke}
                opacity={rule.symbolizer.strokeOpacity}
              />
            )}
            {rule.symbolizer.type === 'point' && (
              <Box position="relative" display="flex" alignItems="center" justifyContent="center">
                {/* Halo effect */}
                {rule.symbolizer.haloColor && rule.symbolizer.haloRadius && rule.symbolizer.haloRadius > 0 && (
                  <Box
                    position="absolute"
                    w={`${(rule.symbolizer.pointSize || 8) + rule.symbolizer.haloRadius * 2}px`}
                    h={`${(rule.symbolizer.pointSize || 8) + rule.symbolizer.haloRadius * 2}px`}
                    borderRadius={rule.symbolizer.pointShape === 'circle' ? '50%' : rule.symbolizer.pointShape === 'triangle' ? '0' : 'sm'}
                    bg={rule.symbolizer.haloColor}
                    opacity={0.4}
                    transform={rule.symbolizer.rotation ? `rotate(${rule.symbolizer.rotation}deg)` : undefined}
                    style={{
                      clipPath: rule.symbolizer.pointShape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                               rule.symbolizer.pointShape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                               rule.symbolizer.pointShape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                               rule.symbolizer.pointShape === 'x' ? 'polygon(10% 0%, 50% 40%, 90% 0%, 100% 10%, 60% 50%, 100% 90%, 90% 100%, 50% 60%, 10% 100%, 0% 90%, 40% 50%, 0% 10%)' :
                               undefined,
                    }}
                  />
                )}
                {/* Main point symbol */}
                <Box
                  position="relative"
                  w={`${rule.symbolizer.pointSize || 8}px`}
                  h={`${rule.symbolizer.pointSize || 8}px`}
                  borderRadius={rule.symbolizer.pointShape === 'circle' ? '50%' : rule.symbolizer.pointShape === 'triangle' ? '0' : 'sm'}
                  bg={rule.symbolizer.fill}
                  opacity={rule.symbolizer.fillOpacity}
                  border={`${rule.symbolizer.strokeWidth}px solid ${rule.symbolizer.stroke}`}
                  transform={rule.symbolizer.rotation ? `rotate(${rule.symbolizer.rotation}deg)` : undefined}
                  style={{
                    clipPath: rule.symbolizer.pointShape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                             rule.symbolizer.pointShape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                             rule.symbolizer.pointShape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                             rule.symbolizer.pointShape === 'x' ? 'polygon(10% 0%, 50% 40%, 90% 0%, 100% 10%, 60% 50%, 100% 90%, 90% 100%, 50% 60%, 10% 100%, 0% 90%, 40% 50%, 0% 10%)' :
                             undefined,
                  }}
                />
              </Box>
            )}
          </Box>
        </Box>
      </VStack>
    </Box>
  )
}
