### Discovery flow through bootstrapping
- Started a slimmed down version of an Ethereum node, just enabling discovery discv4 services on a certain UDP port. This node will act as the Bootstrap node.
    - In devp2p, the Node Identifier is formed out fo the secp256k public key of the predefined key pair. Public key is used for discovery (NodeID) and message/peer validation, and private key for signing.  
    - It starts and inits an in-memory DB for the DHT
    - Discovery services are handled within `discover.UDPv4.readloop()` function.
    - Network layer uses RLP as the transport protocol. Every type of message (ping/pong/...) is encoded in such way. 
- Node 1 ( p2p.Server ) starts listening on local address, using the bootnode for initial bootstrapping. In addition it can initialize any protocol defined.
    - Initializes local DB, tcp services, handshake configuration using local keys and NAT services, if defined.
    - Starts discovery services (similar as the bootstrap node above), with the difference of triggering a random walk to initialize the local DHT.
    - During the random lookup, Node 1 generates a random ID (targetID), and sends `findPeer` request to the known peers from the local table asking for more info about that `targetID`. In the process the known peers wil return another list of peers to keep asking or the local ENR record if found. Since it is a random walk, would be very rare to find the actual peer, which is fine, the main goal is to populate the DHT with neighbor data.
    - Since we start Node 1 first, it will not find any peers to connect to (aside of asking the bootnode).
- Start Node 2 (same as above)
    -  The main difference here, is that during the random walk, bootnode eventually will return the nodeID of Node 1 as part of the Neighbor/v4 structure and at that point is when Node 2 can open a connection to Node 1, and the protocol negotiation starts.
- As soon as Node 1 and Node 2 have open a connection, and negotiated the protocols, each protocol handler will start. In this test scenario, a sequential message interchange will be initiated, in addition to the reverse string logic for each message exchanged.
    - Each peer sends a message containing a sequence number and a string to the connected remote peer every 2 seconds.
    - Each peer listens to incoming messages. There are two types of messages:
        - Request to reverse a string
        - Response with the string reversed
    - Ideally you set up a different struct for each using RLP for each message type. So a simple switch can identify what logic to trigger. For simplicity it only checks on the current sequence number.
    
 
