import { useState } from 'react'
import {
  Box,
  VStack,
  Heading,
  Text,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Input,
  Button,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Alert,
  AlertIcon,
  useColorModeValue,
  Link,
  HStack,
} from '@chakra-ui/react'
import * as hostingApi from '../../api/hosting'

interface AuthFlowProps {
  onSuccess: () => void
}

export default function AuthFlow({ onSuccess }: AuthFlowProps) {
  const [tabIndex, setTabIndex] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Login form
  const [loginEmail, setLoginEmail] = useState('')
  const [loginPassword, setLoginPassword] = useState('')

  // Register form
  const [registerEmail, setRegisterEmail] = useState('')
  const [registerPassword, setRegisterPassword] = useState('')
  const [registerConfirmPassword, setRegisterConfirmPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')

  // Password reset
  const [showResetForm, setShowResetForm] = useState(false)
  const [resetEmail, setResetEmail] = useState('')
  const [resetSent, setResetSent] = useState(false)

  const cardBg = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    try {
      await hostingApi.login({ email: loginEmail, password: loginPassword })
      onSuccess()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    if (registerPassword !== registerConfirmPassword) {
      setError('Passwords do not match')
      setIsLoading(false)
      return
    }

    if (registerPassword.length < 8) {
      setError('Password must be at least 8 characters')
      setIsLoading(false)
      return
    }

    try {
      await hostingApi.register({
        email: registerEmail,
        password: registerPassword,
        first_name: firstName,
        last_name: lastName,
      })
      onSuccess()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  const handlePasswordReset = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    try {
      await hostingApi.requestPasswordReset(resetEmail)
      setResetSent(true)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  if (showResetForm) {
    return (
      <Box
        maxW="md"
        mx="auto"
        p={8}
        bg={cardBg}
        borderRadius="lg"
        borderWidth="1px"
        borderColor={borderColor}
      >
        <VStack spacing={6}>
          <Heading size="md">Reset Password</Heading>

          {resetSent ? (
            <>
              <Alert status="success">
                <AlertIcon />
                Password reset email sent. Check your inbox.
              </Alert>
              <Button variant="link" onClick={() => { setShowResetForm(false); setResetSent(false) }}>
                Back to login
              </Button>
            </>
          ) : (
            <form onSubmit={handlePasswordReset} style={{ width: '100%' }}>
              <VStack spacing={4}>
                {error && (
                  <Alert status="error">
                    <AlertIcon />
                    {error}
                  </Alert>
                )}

                <Text color="gray.500">
                  Enter your email address and we'll send you a link to reset your password.
                </Text>

                <FormControl isRequired>
                  <FormLabel>Email</FormLabel>
                  <Input
                    type="email"
                    value={resetEmail}
                    onChange={(e) => setResetEmail(e.target.value)}
                    placeholder="you@example.com"
                  />
                </FormControl>

                <Button
                  type="submit"
                  colorScheme="blue"
                  w="full"
                  isLoading={isLoading}
                >
                  Send Reset Link
                </Button>

                <Button variant="link" onClick={() => setShowResetForm(false)}>
                  Back to login
                </Button>
              </VStack>
            </form>
          )}
        </VStack>
      </Box>
    )
  }

  return (
    <Box
      maxW="md"
      mx="auto"
      p={8}
      bg={cardBg}
      borderRadius="lg"
      borderWidth="1px"
      borderColor={borderColor}
    >
      <VStack spacing={6}>
        <Heading size="md">Sign in to continue</Heading>
        <Text color="gray.500" textAlign="center">
          Create an account or sign in to complete your order
        </Text>

        {error && (
          <Alert status="error" w="full">
            <AlertIcon />
            {error}
          </Alert>
        )}

        <Tabs
          index={tabIndex}
          onChange={setTabIndex}
          variant="enclosed"
          w="full"
          isFitted
        >
          <TabList>
            <Tab>Sign In</Tab>
            <Tab>Create Account</Tab>
          </TabList>

          <TabPanels>
            {/* Login Panel */}
            <TabPanel px={0}>
              <form onSubmit={handleLogin}>
                <VStack spacing={4}>
                  <FormControl isRequired>
                    <FormLabel>Email</FormLabel>
                    <Input
                      type="email"
                      value={loginEmail}
                      onChange={(e) => setLoginEmail(e.target.value)}
                      placeholder="you@example.com"
                    />
                  </FormControl>

                  <FormControl isRequired>
                    <FormLabel>Password</FormLabel>
                    <Input
                      type="password"
                      value={loginPassword}
                      onChange={(e) => setLoginPassword(e.target.value)}
                      placeholder="Your password"
                    />
                  </FormControl>

                  <Button
                    type="submit"
                    colorScheme="blue"
                    w="full"
                    isLoading={isLoading}
                  >
                    Sign In
                  </Button>

                  <Link
                    color="blue.500"
                    fontSize="sm"
                    onClick={() => setShowResetForm(true)}
                  >
                    Forgot your password?
                  </Link>
                </VStack>
              </form>
            </TabPanel>

            {/* Register Panel */}
            <TabPanel px={0}>
              <form onSubmit={handleRegister}>
                <VStack spacing={4}>
                  <HStack w="full">
                    <FormControl>
                      <FormLabel>First Name</FormLabel>
                      <Input
                        value={firstName}
                        onChange={(e) => setFirstName(e.target.value)}
                        placeholder="John"
                      />
                    </FormControl>

                    <FormControl>
                      <FormLabel>Last Name</FormLabel>
                      <Input
                        value={lastName}
                        onChange={(e) => setLastName(e.target.value)}
                        placeholder="Doe"
                      />
                    </FormControl>
                  </HStack>

                  <FormControl isRequired>
                    <FormLabel>Email</FormLabel>
                    <Input
                      type="email"
                      value={registerEmail}
                      onChange={(e) => setRegisterEmail(e.target.value)}
                      placeholder="you@example.com"
                    />
                  </FormControl>

                  <FormControl isRequired>
                    <FormLabel>Password</FormLabel>
                    <Input
                      type="password"
                      value={registerPassword}
                      onChange={(e) => setRegisterPassword(e.target.value)}
                      placeholder="Minimum 8 characters"
                    />
                  </FormControl>

                  <FormControl isRequired isInvalid={registerConfirmPassword !== '' && registerPassword !== registerConfirmPassword}>
                    <FormLabel>Confirm Password</FormLabel>
                    <Input
                      type="password"
                      value={registerConfirmPassword}
                      onChange={(e) => setRegisterConfirmPassword(e.target.value)}
                      placeholder="Confirm your password"
                    />
                    {registerConfirmPassword !== '' && registerPassword !== registerConfirmPassword && (
                      <FormErrorMessage>Passwords do not match</FormErrorMessage>
                    )}
                  </FormControl>

                  <Button
                    type="submit"
                    colorScheme="blue"
                    w="full"
                    isLoading={isLoading}
                  >
                    Create Account
                  </Button>

                  <Text fontSize="xs" color="gray.500" textAlign="center">
                    By creating an account, you agree to our Terms of Service and Privacy Policy.
                  </Text>
                </VStack>
              </form>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>
    </Box>
  )
}
