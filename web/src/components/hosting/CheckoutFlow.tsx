import { useState, useEffect } from 'react'
import {
  Box,
  VStack,
  HStack,
  Heading,
  Text,
  Card,
  CardBody,
  Button,
  Alert,
  AlertIcon,
  Spinner,
  useColorModeValue,
  Divider,
  Badge,
  Icon,
  RadioGroup,
  Radio,
  Stack,
} from '@chakra-ui/react'
import { FaCreditCard, FaCheckCircle, FaExternalLinkAlt } from 'react-icons/fa'
import { SiStripe } from 'react-icons/si'
import * as hostingApi from '../../api/hosting'
import type { Product, Package, Cluster, PaymentConfig, SalesOrder } from '../../api/hosting'

interface CheckoutFlowProps {
  product: Product
  pkg: Package
  cluster: Cluster
  appName: string
  billingCycle: 'monthly' | 'yearly'
  onComplete: () => void
}

type CheckoutStep = 'review' | 'payment' | 'processing' | 'success'

export default function CheckoutFlow({
  product,
  pkg,
  cluster,
  appName,
  billingCycle,
  onComplete,
}: CheckoutFlowProps) {
  const [step, setStep] = useState<CheckoutStep>('review')
  const [paymentConfig, setPaymentConfig] = useState<PaymentConfig | null>(null)
  const [selectedProvider, setSelectedProvider] = useState<'stripe' | 'paystack'>('stripe')
  const [order, setOrder] = useState<SalesOrder | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const cardBg = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  // Load payment config
  useEffect(() => {
    loadPaymentConfig()
  }, [])

  const loadPaymentConfig = async () => {
    try {
      const config = await hostingApi.getPaymentConfig()
      setPaymentConfig(config)
      if (config.providers.length > 0) {
        setSelectedProvider(config.providers[0] as 'stripe' | 'paystack')
      }
    } catch (err) {
      setError((err as Error).message)
    }
  }

  const price = billingCycle === 'yearly' ? pkg.price_yearly : pkg.price_monthly

  const handleCreateOrder = async () => {
    setIsLoading(true)
    setError(null)

    try {
      const newOrder = await hostingApi.createOrder({
        package_id: pkg.id,
        cluster_id: cluster.id,
        app_name: appName,
        billing_cycle: billingCycle,
      })
      setOrder(newOrder)
      setStep('payment')
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCheckout = async () => {
    if (!order) return

    setIsLoading(true)
    setError(null)

    try {
      const baseUrl = window.location.origin
      const checkoutResponse = await hostingApi.checkout(order.id, selectedProvider, {
        success_url: `${baseUrl}/hosting/order/${order.id}/success`,
        cancel_url: `${baseUrl}/hosting/order/${order.id}/cancel`,
      })

      // Redirect to payment provider
      window.location.href = checkoutResponse.checkout_url
    } catch (err) {
      setError((err as Error).message)
      setIsLoading(false)
    }
  }

  const renderReviewStep = () => (
    <VStack spacing={6} align="stretch">
      <Heading size="md">Review Your Order</Heading>

      <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
        <CardBody>
          <VStack spacing={4} align="stretch">
            <HStack justify="space-between">
              <Text fontWeight="bold">Product</Text>
              <Text>{product.name}</Text>
            </HStack>
            <HStack justify="space-between">
              <Text fontWeight="bold">Plan</Text>
              <Text>{pkg.name}</Text>
            </HStack>
            <HStack justify="space-between">
              <Text fontWeight="bold">Region</Text>
              <Text>{cluster.name}</Text>
            </HStack>
            <HStack justify="space-between">
              <Text fontWeight="bold">Instance Name</Text>
              <Text>{appName}</Text>
            </HStack>
            <HStack justify="space-between">
              <Text fontWeight="bold">URL</Text>
              <Text color="blue.500">{appName}.{cluster.domain}</Text>
            </HStack>

            <Divider />

            <HStack justify="space-between">
              <Text fontWeight="bold">Billing</Text>
              <Badge colorScheme={billingCycle === 'yearly' ? 'green' : 'blue'}>
                {billingCycle === 'yearly' ? 'Yearly' : 'Monthly'}
              </Badge>
            </HStack>

            <Divider />

            <HStack justify="space-between" fontSize="lg">
              <Text fontWeight="bold">Total</Text>
              <VStack align="end" spacing={0}>
                <Text fontWeight="bold">{hostingApi.formatPrice(price)}</Text>
                <Text fontSize="sm" color="gray.500">
                  /{billingCycle === 'yearly' ? 'year' : 'month'}
                </Text>
              </VStack>
            </HStack>

            {billingCycle === 'yearly' && (
              <Alert status="success" borderRadius="md">
                <AlertIcon />
                You save {hostingApi.formatPrice(pkg.price_monthly * 12 - pkg.price_yearly)} per year!
              </Alert>
            )}
          </VStack>
        </CardBody>
      </Card>

      {error && (
        <Alert status="error">
          <AlertIcon />
          {error}
        </Alert>
      )}

      <Button
        colorScheme="blue"
        size="lg"
        onClick={handleCreateOrder}
        isLoading={isLoading}
      >
        Continue to Payment
      </Button>

      <Text fontSize="xs" color="gray.500" textAlign="center">
        Your instance will be deployed automatically after payment is confirmed.
      </Text>
    </VStack>
  )

  const renderPaymentStep = () => (
    <VStack spacing={6} align="stretch">
      <Heading size="md">Select Payment Method</Heading>

      {/* Order Summary */}
      <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
        <CardBody>
          <HStack justify="space-between">
            <VStack align="start" spacing={0}>
              <Text fontWeight="bold">{product.name} - {pkg.name}</Text>
              <Text fontSize="sm" color="gray.500">{appName}.{cluster.domain}</Text>
            </VStack>
            <Text fontWeight="bold">{hostingApi.formatPrice(price)}</Text>
          </HStack>
        </CardBody>
      </Card>

      {/* Payment Provider Selection */}
      {paymentConfig && paymentConfig.providers.length > 0 ? (
        <RadioGroup value={selectedProvider} onChange={(v) => setSelectedProvider(v as 'stripe' | 'paystack')}>
          <Stack spacing={4}>
            {paymentConfig.providers.includes('stripe') && (
              <Card
                cursor="pointer"
                onClick={() => setSelectedProvider('stripe')}
                bg={selectedProvider === 'stripe' ? 'blue.50' : cardBg}
                borderWidth={selectedProvider === 'stripe' ? '2px' : '1px'}
                borderColor={selectedProvider === 'stripe' ? 'blue.500' : borderColor}
                _hover={{ borderColor: 'blue.500' }}
              >
                <CardBody>
                  <HStack justify="space-between">
                    <HStack spacing={4}>
                      <Radio value="stripe" colorScheme="blue" />
                      <Icon as={SiStripe} boxSize={8} color="purple.600" />
                      <VStack align="start" spacing={0}>
                        <Text fontWeight="bold">Stripe</Text>
                        <Text fontSize="sm" color="gray.500">Credit/Debit Card</Text>
                      </VStack>
                    </HStack>
                    <HStack spacing={2}>
                      <Icon as={FaCreditCard} color="gray.500" />
                    </HStack>
                  </HStack>
                </CardBody>
              </Card>
            )}

            {paymentConfig.providers.includes('paystack') && (
              <Card
                cursor="pointer"
                onClick={() => setSelectedProvider('paystack')}
                bg={selectedProvider === 'paystack' ? 'blue.50' : cardBg}
                borderWidth={selectedProvider === 'paystack' ? '2px' : '1px'}
                borderColor={selectedProvider === 'paystack' ? 'blue.500' : borderColor}
                _hover={{ borderColor: 'blue.500' }}
              >
                <CardBody>
                  <HStack justify="space-between">
                    <HStack spacing={4}>
                      <Radio value="paystack" colorScheme="blue" />
                      <Box
                        w={8}
                        h={8}
                        bg="teal.500"
                        borderRadius="md"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                        color="white"
                        fontWeight="bold"
                        fontSize="xs"
                      >
                        PS
                      </Box>
                      <VStack align="start" spacing={0}>
                        <Text fontWeight="bold">Paystack</Text>
                        <Text fontSize="sm" color="gray.500">Cards, Bank Transfer, Mobile Money</Text>
                      </VStack>
                    </HStack>
                    <HStack spacing={2}>
                      <Icon as={FaCreditCard} color="gray.500" />
                    </HStack>
                  </HStack>
                </CardBody>
              </Card>
            )}
          </Stack>
        </RadioGroup>
      ) : (
        <Alert status="warning">
          <AlertIcon />
          No payment providers configured. Please contact support.
        </Alert>
      )}

      {error && (
        <Alert status="error">
          <AlertIcon />
          {error}
        </Alert>
      )}

      <Button
        colorScheme="blue"
        size="lg"
        onClick={handleCheckout}
        isLoading={isLoading}
        isDisabled={!paymentConfig?.providers.length}
        rightIcon={<FaExternalLinkAlt />}
      >
        Pay with {selectedProvider === 'stripe' ? 'Stripe' : 'Paystack'}
      </Button>

      <Text fontSize="xs" color="gray.500" textAlign="center">
        You will be redirected to {selectedProvider === 'stripe' ? 'Stripe' : 'Paystack'} to complete your payment securely.
      </Text>
    </VStack>
  )

  const renderProcessingStep = () => (
    <VStack spacing={6} py={12}>
      <Spinner size="xl" color="blue.500" />
      <Heading size="md">Processing Payment...</Heading>
      <Text color="gray.500">Please wait while we confirm your payment.</Text>
    </VStack>
  )

  const renderSuccessStep = () => (
    <VStack spacing={6} py={12}>
      <Icon as={FaCheckCircle} boxSize={16} color="green.500" />
      <Heading size="md">Payment Successful!</Heading>
      <Text color="gray.500" textAlign="center">
        Your instance is being deployed. You will receive an email when it's ready.
      </Text>

      <Card bg={cardBg} borderWidth="1px" borderColor={borderColor} w="full">
        <CardBody>
          <VStack spacing={2} align="start">
            <HStack justify="space-between" w="full">
              <Text fontWeight="bold">Order ID</Text>
              <Text fontFamily="mono">{order?.id}</Text>
            </HStack>
            <HStack justify="space-between" w="full">
              <Text fontWeight="bold">Instance URL</Text>
              <Text color="blue.500">{appName}.{cluster.domain}</Text>
            </HStack>
            <HStack justify="space-between" w="full">
              <Text fontWeight="bold">Status</Text>
              <Badge colorScheme="yellow">Deploying</Badge>
            </HStack>
          </VStack>
        </CardBody>
      </Card>

      <Button colorScheme="blue" onClick={onComplete}>
        View My Instances
      </Button>
    </VStack>
  )

  const renderContent = () => {
    switch (step) {
      case 'review':
        return renderReviewStep()
      case 'payment':
        return renderPaymentStep()
      case 'processing':
        return renderProcessingStep()
      case 'success':
        return renderSuccessStep()
      default:
        return null
    }
  }

  return (
    <Box maxW="lg" mx="auto">
      {renderContent()}
    </Box>
  )
}
