import { useState, useEffect } from 'react'
import {
  Box,
  VStack,
  HStack,
  Heading,
  Text,
  SimpleGrid,
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  Button,
  Badge,
  Switch,
  FormControl,
  FormLabel,
  Icon,
  List,
  ListItem,
  ListIcon,
  Divider,
  useColorModeValue,
  Spinner,
  Alert,
  AlertIcon,
  IconButton,
} from '@chakra-ui/react'
import {
  FaServer,
  FaDatabase,
  FaGlobe,
  FaCheck,
  FaStar,
  FaArrowRight,
  FaArrowLeft,
} from 'react-icons/fa'
import * as hostingApi from '../../api/hosting'
import type { Product, Package, Cluster } from '../../api/hosting'
import AuthFlow from './AuthFlow'
import CheckoutFlow from './CheckoutFlow'

type ShopStep = 'browse' | 'auth' | 'configure' | 'checkout'

export default function ShopPanel() {
  const [step, setStep] = useState<ShopStep>('browse')
  const [products, setProducts] = useState<Product[]>([])
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)
  const [packages, setPackages] = useState<Package[]>([])
  const [selectedPackage, setSelectedPackage] = useState<Package | null>(null)
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [selectedCluster, setSelectedCluster] = useState<Cluster | null>(null)
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly')
  const [appName, setAppName] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const cardBg = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const popularBg = useColorModeValue('blue.50', 'blue.900')

  // Load products on mount
  useEffect(() => {
    loadProducts()
  }, [])

  // Load packages when product is selected
  useEffect(() => {
    if (selectedProduct) {
      loadPackages(selectedProduct.slug)
    }
  }, [selectedProduct])

  // Load clusters when moving to configure step
  useEffect(() => {
    if (step === 'configure') {
      loadClusters()
    }
  }, [step])

  const loadProducts = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await hostingApi.getProducts()
      setProducts(data)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  const loadPackages = async (productSlug: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await hostingApi.getProductPackages(productSlug)
      setPackages(data)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  const loadClusters = async () => {
    try {
      const data = await hostingApi.getClusters()
      setClusters(data.filter(c => c.is_active))
      if (data.length > 0) {
        setSelectedCluster(data[0])
      }
    } catch (err) {
      setError((err as Error).message)
    }
  }

  const getProductIcon = (slug: string) => {
    switch (slug) {
      case 'geoserver':
        return FaServer
      case 'geonode':
        return FaGlobe
      case 'postgis':
        return FaDatabase
      default:
        return FaServer
    }
  }

  const handleSelectProduct = (product: Product) => {
    setSelectedProduct(product)
    setSelectedPackage(null)
  }

  const handleSelectPackage = (pkg: Package) => {
    setSelectedPackage(pkg)
    // Check if user is authenticated
    if (hostingApi.isAuthenticated()) {
      setStep('configure')
    } else {
      setStep('auth')
    }
  }

  const handleAuthSuccess = () => {
    setStep('configure')
  }

  const handleConfigureComplete = () => {
    setStep('checkout')
  }

  const handleBack = () => {
    switch (step) {
      case 'auth':
        setStep('browse')
        setSelectedPackage(null)
        break
      case 'configure':
        setStep('browse')
        break
      case 'checkout':
        setStep('configure')
        break
    }
  }

  const renderBrowseStep = () => (
    <VStack spacing={8} align="stretch">
      {/* Hero Section */}
      <Box textAlign="center" py={8}>
        <Heading size="xl" mb={4}>Geospatial Hosting</Heading>
        <Text fontSize="lg" color="gray.500" maxW="2xl" mx="auto">
          Deploy managed GeoServer, GeoNode, and PostGIS instances in minutes.
          Fully managed infrastructure with automatic backups and monitoring.
        </Text>
      </Box>

      {/* Product Selection */}
      {!selectedProduct ? (
        <>
          <Heading size="md">Choose Your Product</Heading>
          {isLoading ? (
            <Box textAlign="center" py={8}>
              <Spinner size="lg" />
            </Box>
          ) : error ? (
            <Alert status="error">
              <AlertIcon />
              {error}
            </Alert>
          ) : (
            <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6}>
              {products.map((product) => (
                <Card
                  key={product.id}
                  cursor="pointer"
                  onClick={() => handleSelectProduct(product)}
                  bg={cardBg}
                  borderWidth="1px"
                  borderColor={borderColor}
                  _hover={{ transform: 'translateY(-4px)', shadow: 'lg' }}
                  transition="all 0.2s"
                >
                  <CardHeader textAlign="center">
                    <Icon
                      as={getProductIcon(product.slug)}
                      boxSize={12}
                      color="blue.500"
                      mb={4}
                    />
                    <Heading size="md">{product.name}</Heading>
                  </CardHeader>
                  <CardBody>
                    <Text color="gray.500">{product.description}</Text>
                  </CardBody>
                  <CardFooter justifyContent="center">
                    <Button rightIcon={<FaArrowRight />} colorScheme="blue" variant="ghost">
                      View Plans
                    </Button>
                  </CardFooter>
                </Card>
              ))}
            </SimpleGrid>
          )}
        </>
      ) : (
        <>
          {/* Back button and product info */}
          <HStack>
            <IconButton
              aria-label="Back to products"
              icon={<FaArrowLeft />}
              onClick={() => setSelectedProduct(null)}
              variant="ghost"
            />
            <Icon as={getProductIcon(selectedProduct.slug)} boxSize={6} color="blue.500" />
            <Heading size="md">{selectedProduct.name} Plans</Heading>
          </HStack>

          {/* Billing Cycle Toggle */}
          <HStack justify="center" spacing={4}>
            <Text color={billingCycle === 'monthly' ? 'blue.500' : 'gray.500'}>Monthly</Text>
            <Switch
              isChecked={billingCycle === 'yearly'}
              onChange={(e) => setBillingCycle(e.target.checked ? 'yearly' : 'monthly')}
              colorScheme="blue"
              size="lg"
            />
            <HStack>
              <Text color={billingCycle === 'yearly' ? 'blue.500' : 'gray.500'}>Yearly</Text>
              {billingCycle === 'yearly' && (
                <Badge colorScheme="green">Save up to 20%</Badge>
              )}
            </HStack>
          </HStack>

          {/* Package Cards */}
          {isLoading ? (
            <Box textAlign="center" py={8}>
              <Spinner size="lg" />
            </Box>
          ) : (
            <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6}>
              {packages.map((pkg) => (
                <Card
                  key={pkg.id}
                  bg={pkg.is_popular ? popularBg : cardBg}
                  borderWidth={pkg.is_popular ? '2px' : '1px'}
                  borderColor={pkg.is_popular ? 'blue.500' : borderColor}
                  position="relative"
                >
                  {pkg.is_popular && (
                    <Badge
                      position="absolute"
                      top="-3"
                      left="50%"
                      transform="translateX(-50%)"
                      colorScheme="blue"
                      px={3}
                      py={1}
                    >
                      <HStack spacing={1}>
                        <FaStar />
                        <Text>Most Popular</Text>
                      </HStack>
                    </Badge>
                  )}
                  <CardHeader textAlign="center" pt={pkg.is_popular ? 8 : 4}>
                    <Heading size="md">{pkg.name}</Heading>
                  </CardHeader>
                  <CardBody>
                    <VStack spacing={4}>
                      {/* Price */}
                      <Box textAlign="center">
                        <Text fontSize="4xl" fontWeight="bold">
                          {hostingApi.formatPrice(
                            billingCycle === 'yearly' ? pkg.price_yearly / 12 : pkg.price_monthly
                          )}
                        </Text>
                        <Text color="gray.500">
                          per month{billingCycle === 'yearly' && ', billed yearly'}
                        </Text>
                        {billingCycle === 'yearly' && (
                          <Badge colorScheme="green" mt={1}>
                            Save {hostingApi.calculateYearlySavings(pkg.price_monthly, pkg.price_yearly)}%
                          </Badge>
                        )}
                      </Box>

                      <Divider />

                      {/* Features */}
                      <List spacing={2} w="full">
                        {pkg.features?.map((feature, idx) => (
                          <ListItem key={idx}>
                            <ListIcon as={FaCheck} color="green.500" />
                            {feature}
                          </ListItem>
                        ))}
                        {pkg.cpu_limit && (
                          <ListItem>
                            <ListIcon as={FaCheck} color="green.500" />
                            {pkg.cpu_limit} CPU
                          </ListItem>
                        )}
                        {pkg.memory_limit && (
                          <ListItem>
                            <ListIcon as={FaCheck} color="green.500" />
                            {pkg.memory_limit} Memory
                          </ListItem>
                        )}
                        {pkg.storage_limit && (
                          <ListItem>
                            <ListIcon as={FaCheck} color="green.500" />
                            {pkg.storage_limit} Storage
                          </ListItem>
                        )}
                      </List>
                    </VStack>
                  </CardBody>
                  <CardFooter>
                    <Button
                      w="full"
                      colorScheme="blue"
                      variant={pkg.is_popular ? 'solid' : 'outline'}
                      onClick={() => handleSelectPackage(pkg)}
                    >
                      Select Plan
                    </Button>
                  </CardFooter>
                </Card>
              ))}
            </SimpleGrid>
          )}
        </>
      )}
    </VStack>
  )

  const renderContent = () => {
    switch (step) {
      case 'browse':
        return renderBrowseStep()
      case 'auth':
        return (
          <VStack spacing={6}>
            <IconButton
              aria-label="Back"
              icon={<FaArrowLeft />}
              onClick={handleBack}
              alignSelf="flex-start"
              variant="ghost"
            />
            <AuthFlow onSuccess={handleAuthSuccess} />
          </VStack>
        )
      case 'configure':
        return (
          <VStack spacing={6} align="stretch">
            <HStack>
              <IconButton
                aria-label="Back"
                icon={<FaArrowLeft />}
                onClick={handleBack}
                variant="ghost"
              />
              <Heading size="md">Configure Your Instance</Heading>
            </HStack>

            {/* Selected Package Summary */}
            {selectedPackage && (
              <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
                <CardBody>
                  <HStack justify="space-between">
                    <VStack align="start" spacing={1}>
                      <Text fontWeight="bold">
                        {selectedProduct?.name} - {selectedPackage.name}
                      </Text>
                      <Text color="gray.500">
                        {hostingApi.formatPrice(
                          billingCycle === 'yearly' ? selectedPackage.price_yearly : selectedPackage.price_monthly
                        )}
                        /{billingCycle === 'yearly' ? 'year' : 'month'}
                      </Text>
                    </VStack>
                    <Button size="sm" variant="outline" onClick={() => setStep('browse')}>
                      Change
                    </Button>
                  </HStack>
                </CardBody>
              </Card>
            )}

            {/* Cluster Selection */}
            <FormControl>
              <FormLabel>Select Region</FormLabel>
              <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                {clusters.map((cluster) => (
                  <Card
                    key={cluster.id}
                    cursor="pointer"
                    onClick={() => setSelectedCluster(cluster)}
                    bg={selectedCluster?.id === cluster.id ? popularBg : cardBg}
                    borderWidth={selectedCluster?.id === cluster.id ? '2px' : '1px'}
                    borderColor={selectedCluster?.id === cluster.id ? 'blue.500' : borderColor}
                    _hover={{ borderColor: 'blue.500' }}
                  >
                    <CardBody>
                      <HStack justify="space-between">
                        <VStack align="start" spacing={0}>
                          <Text fontWeight="bold">{cluster.name}</Text>
                          <Text fontSize="sm" color="gray.500">{cluster.region}</Text>
                        </VStack>
                        {selectedCluster?.id === cluster.id && (
                          <Icon as={FaCheck} color="blue.500" />
                        )}
                      </HStack>
                    </CardBody>
                  </Card>
                ))}
              </SimpleGrid>
            </FormControl>

            {/* App Name */}
            <FormControl>
              <FormLabel>Instance Name</FormLabel>
              <input
                type="text"
                value={appName}
                onChange={(e) => setAppName(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                placeholder="my-geoserver"
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: '6px',
                  border: '1px solid var(--chakra-colors-gray-300)',
                  fontSize: '16px',
                }}
              />
              <Text fontSize="sm" color="gray.500" mt={1}>
                Your instance will be available at: {appName || 'my-instance'}.{selectedCluster?.domain || 'example.com'}
              </Text>
            </FormControl>

            <Button
              colorScheme="blue"
              size="lg"
              onClick={handleConfigureComplete}
              isDisabled={!appName || !selectedCluster}
              rightIcon={<FaArrowRight />}
            >
              Continue to Checkout
            </Button>
          </VStack>
        )
      case 'checkout':
        return (
          <VStack spacing={6} align="stretch">
            <IconButton
              aria-label="Back"
              icon={<FaArrowLeft />}
              onClick={handleBack}
              alignSelf="flex-start"
              variant="ghost"
            />
            <CheckoutFlow
              product={selectedProduct!}
              pkg={selectedPackage!}
              cluster={selectedCluster!}
              appName={appName}
              billingCycle={billingCycle}
              onComplete={() => {
                // TODO: Refresh instances and navigate to instance view
                setStep('browse')
              }}
            />
          </VStack>
        )
      default:
        return null
    }
  }

  return (
    <Box p={6} h="full" overflowY="auto">
      {renderContent()}

      {/* Footer */}
      <Box textAlign="center" mt={12} pt={6} borderTopWidth="1px" borderColor={borderColor}>
        <Text fontSize="sm" color="gray.500">
          Made with love by{' '}
          <a href="https://kartoza.com" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--chakra-colors-blue-500)' }}>
            Kartoza
          </a>
          {' | '}
          <a href="https://github.com/sponsors/kartoza" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--chakra-colors-blue-500)' }}>
            Donate
          </a>
          {' | '}
          <a href="https://github.com/kartoza/kartoza-cloudbench" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--chakra-colors-blue-500)' }}>
            GitHub
          </a>
        </Text>
      </Box>
    </Box>
  )
}
