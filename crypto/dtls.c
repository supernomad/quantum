#include "dtls.h"

// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

unsigned char cookie_secret[COOKIE_SECRET_LENGTH];

void init_dtls() {
    // Need to init the ssl libs and load error strings.
    SSL_library_init();
    SSL_load_error_strings();
    OpenSSL_add_all_algorithms();

    // Need a random secret for the cookie generation.
    RAND_bytes(cookie_secret, COOKIE_SECRET_LENGTH);
}

void destroy_dtls() {
    // Need to cleanup after openssl.
    ENGINE_cleanup();
    CONF_modules_unload(1);
    ERR_free_strings();
    EVP_cleanup();
    CRYPTO_cleanup_all_ex_data();
}

int _parse_sock_addr(int is_v6, const char* addr, int port, struct sockaddr* sa) {
    int size = -1;

    if (is_v6) {
        struct sockaddr_in6 local_addr;
        memset(&local_addr, '\0', sizeof(local_addr));

        local_addr.sin6_family = AF_INET6;
        local_addr.sin6_port = htons(port);
        inet_pton(AF_INET6, addr, &local_addr.sin6_addr);

        size = sizeof local_addr;
        memcpy(sa, (struct sockaddr*)&local_addr, size);
    } else {
        struct sockaddr_in local_addr;
        memset(&local_addr, '\0', sizeof(local_addr));

        local_addr.sin_family = AF_INET;
        local_addr.sin_port = htons(port);
        inet_pton(AF_INET, addr, &local_addr.sin_addr);

        size = sizeof local_addr;
        memcpy(sa, (struct sockaddr*)&local_addr, size);
    }

    return size;
}

int _verify_peer_cb(int ok, X509_STORE_CTX* ctx) {
    // Rely on the built in openssl verification process.
    return ok;
}


int cookie_initialized=0;

int _generate_cookie_cb(SSL* ssl, unsigned char* cookie, unsigned int* cookie_len) {
    unsigned char *buffer, result[EVP_MAX_MD_SIZE];
    unsigned int length = 0, resultlength;
    union {
        struct sockaddr_storage ss;
        struct sockaddr_in6 s6;
        struct sockaddr_in s4;
    } peer;

    // Read peer information from ssl object
    BIO_dgram_get_peer(SSL_get_rbio(ssl), &peer);

    // Create buffer with peer's address and port
    length = 0;
    switch (peer.ss.ss_family) {
        case AF_INET:
            length += sizeof(struct in_addr);
            break;
        case AF_INET6:
            length += sizeof(struct in6_addr);
            break;
        default:
            return 0;
    }

    length += sizeof(in_port_t);
    buffer = (unsigned char*) OPENSSL_malloc(length);

    if (buffer == NULL) {
        return 0;
    }

    switch (peer.ss.ss_family) {
        case AF_INET:
            memcpy(buffer, &peer.s4.sin_port, sizeof(in_port_t));
            memcpy(buffer + sizeof(peer.s4.sin_port), &peer.s4.sin_addr, sizeof(struct in_addr));
            break;
        case AF_INET6:
            memcpy(buffer, &peer.s6.sin6_port, sizeof(in_port_t));
            memcpy(buffer + sizeof(in_port_t), &peer.s6.sin6_addr, sizeof(struct in6_addr));
            break;
        default:
            return 0;
    }

    // Calculate HMAC of buffer using the secret
    HMAC(EVP_sha1(), (const void*) cookie_secret, COOKIE_SECRET_LENGTH, (const unsigned char*) buffer, length, result, &resultlength);
    OPENSSL_free(buffer);

    memcpy(cookie, result, resultlength);
    *cookie_len = resultlength;

    return 1;
}

int _verify_cookie_cb(SSL* ssl, const unsigned char* cookie, unsigned int cookie_len) {
    unsigned char *buffer, result[EVP_MAX_MD_SIZE];
    unsigned int length = 0, resultlength;
    union {
        struct sockaddr_storage ss;
        struct sockaddr_in6 s6;
        struct sockaddr_in s4;
    } peer;

    // Read peer information from ssl object
    BIO_dgram_get_peer(SSL_get_rbio(ssl), &peer);

    // Create buffer with peer's address and port
    length = 0;
    switch (peer.ss.ss_family) {
        case AF_INET:
            length += sizeof(struct in_addr);
            break;
        case AF_INET6:
            length += sizeof(struct in6_addr);
            break;
        default:
            return 0;
    }
    length += sizeof(in_port_t);
    buffer = (unsigned char*) OPENSSL_malloc(length);

    if (buffer == NULL) {
        return 0;
    }

    switch (peer.ss.ss_family) {
        case AF_INET:
            memcpy(buffer,
                   &peer.s4.sin_port,
                   sizeof(in_port_t));
            memcpy(buffer + sizeof(in_port_t),
                   &peer.s4.sin_addr,
                   sizeof(struct in_addr));
            break;
        case AF_INET6:
            memcpy(buffer,
                   &peer.s6.sin6_port,
                   sizeof(in_port_t));
            memcpy(buffer + sizeof(in_port_t),
                   &peer.s6.sin6_addr,
                   sizeof(struct in6_addr));
            break;
        default:
            return 0;
    }

    // Calculate HMAC of buffer using the secret
    HMAC(EVP_sha1(), (const void*) cookie_secret, COOKIE_SECRET_LENGTH, (const unsigned char*) buffer, length, result, &resultlength);
    OPENSSL_free(buffer);

    if (cookie_len == resultlength && memcmp(result, cookie, resultlength) == 0) {
        return 1;
    }

    return 0;
}

int _set_common_ssl_parameters(SSL_CTX* ssl_ctx, int verify_peer, char* error) {
    // Set the supported cipher to our hard coded cipher.
    if (!SSL_CTX_set_cipher_list(ssl_ctx, SSL_CIPHER)) {
        strcpy(error, "unable set cipher list to appropriate value");
        return 0;
    }

    // Set the ecdh functionality to enabled.
    if (!SSL_CTX_set_ecdh_auto(ssl_ctx, 1)) {
        strcpy(error, "unable to enable required ecdh functionality");
        return 0;
    }

    // Set the minimum DTLS version accepted to v1.2.
    if (!SSL_CTX_set_min_proto_version(ssl_ctx, DTLS1_2_VERSION)) {
        strcpy(error, "required DTLS version is unsupported");
        return 0;
    }

    // Set the maximum DTLS version accepted to v1.2.
    if (!SSL_CTX_set_max_proto_version(ssl_ctx, DTLS1_2_VERSION)) {
        strcpy(error, "required DTLS version is unsupported");
        return 0;
    }

    SSL_CTX_set_options(ssl_ctx, SSL_OP_NO_QUERY_MTU);

    // Set the read ahead functionality to true.
    SSL_CTX_set_read_ahead(ssl_ctx, 1);

    // Set the caching mode.
    SSL_CTX_set_session_cache_mode(ssl_ctx, SSL_SESS_CACHE_BOTH | SSL_SESS_CACHE_NO_AUTO_CLEAR);

    if(verify_peer) {
        // Set the verification settings for incoming/outgoing certificates, so that verification is fully enforced. SSL_VERIFY_CLIENT_ONCE is ignored on client side connections.
        SSL_CTX_set_verify(ssl_ctx, SSL_VERIFY_PEER | SSL_VERIFY_FAIL_IF_NO_PEER_CERT | SSL_VERIFY_CLIENT_ONCE, _verify_peer_cb);
    } else {
        // Set the verification settings for incoming/outgoing certificates, so that verification is ignored.
        SSL_CTX_set_verify(ssl_ctx, SSL_VERIFY_NONE, _verify_peer_cb);
    }

    // SSL parameters set properly.
    return 1;
}

int _set_ssl_certs(SSL_CTX* ssl_ctx, const char* ca, const char* cert, const char* key, char* error) {
    // Load the ca certificate to use for verification.
    if (!SSL_CTX_load_verify_locations(ssl_ctx, ca, NULL)) {
        strcpy(error, "unable to load the specified CA certificate file");
        return 0;
    }

    // Load the public certificate.
    if (!SSL_CTX_use_certificate_file(ssl_ctx, cert, SSL_FILETYPE_PEM)) {
        strcpy(error, "unable to load the specified public certificate file");
        return 0;
    }

    // Load the private key for the above public certificate.
    if (!SSL_CTX_use_PrivateKey_file(ssl_ctx, key, SSL_FILETYPE_PEM)) {
        strcpy(error, "unable to load the specified private key file");
        return 0;
    }

    // Check the validity of the private key to ensure proper operation.
    if (!SSL_CTX_check_private_key(ssl_ctx)) {
        strcpy(error, "the specified private key file does not match the specified pubcli certificate file");
        return 0;
    }

    // Certificates and keys set properly.
    return 1;
}

Context* init_server_dtls_context(int fd, const char* addr, int port, int use_v6, int verify_peer, const char* ca, const char* cert, const char* key, char* error) {
    Context* ctx = (Context*)malloc(sizeof(Context));

    ctx->use_v6 = use_v6;
    ctx->fd = fd;

    if (ctx->use_v6) {
        ctx->local_addr_len = sizeof(struct sockaddr_in6);
    } else {
        ctx->local_addr_len = sizeof(struct sockaddr_in);
    }

    ctx->local_addr = (struct sockaddr*)malloc(ctx->local_addr_len);
    _parse_sock_addr(ctx->use_v6, addr, port, ctx->local_addr);

    // Create a new DTLS context for the server.
    ctx->ssl_ctx = SSL_CTX_new(DTLS_server_method());
    if (ctx->ssl_ctx == NULL) {
        strcpy(error, "unable to create the server SSL context");
        free_dtls_context(ctx);
        return NULL;
    }

    // Setup the verification and cookie exchange callbacks.
    SSL_CTX_set_cookie_generate_cb(ctx->ssl_ctx, _generate_cookie_cb);
    SSL_CTX_set_cookie_verify_cb(ctx->ssl_ctx, _verify_cookie_cb);

    // Set the common DTLS parameters.
    if (!_set_common_ssl_parameters(ctx->ssl_ctx, verify_peer, error)) {
        free_dtls_context(ctx);
        return NULL;
    }

    if (!_set_ssl_certs(ctx->ssl_ctx, ca, cert, key, error)) {
        free_dtls_context(ctx);
        return NULL;
    }

    return ctx;
}

Context* init_client_dtls_context(const char* addr, int use_v6, int verify_peer, const char* ca, const char* cert, const char* key, char* error) {
    Context* ctx = (Context*)malloc(sizeof(Context));

    ctx->use_v6 = use_v6;
    ctx->fd = -1;

    if (ctx->use_v6) {
        ctx->local_addr_len = sizeof(struct sockaddr_in6);
    } else {
        ctx->local_addr_len = sizeof(struct sockaddr_in);
    }

    ctx->local_addr = (struct sockaddr*)malloc(ctx->local_addr_len);
    _parse_sock_addr(ctx->use_v6, addr, CLIENT_CTX_PORT, ctx->local_addr);

    // Create a new DTLS context for the client.
    ctx->ssl_ctx = SSL_CTX_new(DTLS_client_method());
    if (ctx->ssl_ctx == NULL) {
        strcpy(error, "unable to create the client SSL context");
        free_dtls_context(ctx);
        return NULL;
    }

    // Set the common DTLS parameters.
    if (!_set_common_ssl_parameters(ctx->ssl_ctx, verify_peer, error)) {
        free_dtls_context(ctx);
        return NULL;
    }

    // Set the specified certificate/key for the context.
    if (!_set_ssl_certs(ctx->ssl_ctx, ca, cert, key, error)) {
        free_dtls_context(ctx);
        return NULL;
    }

    return ctx;
}

void free_dtls_context(Context* ctx) {
    if (ctx == NULL) {
        return;
    }

    if (ctx->ssl_ctx != NULL) {
        SSL_CTX_free(ctx->ssl_ctx);
        ctx->ssl_ctx = NULL;
    }

    if (ctx->fd > 0) {
        close(ctx->fd);
        ctx->fd = -1;
    }
}

Session* accept_dtls(Context* ctx, char* error) {
    Session* session = (Session*)malloc(sizeof(Session));

    BIO* bio = BIO_new_dgram(ctx->fd, BIO_NOCLOSE);
    if (bio == NULL) {
        strcpy(error, "unable to create the required SSL BIO object for the new connection");
        free_dtls_session(session);
        return NULL;
    }

    session->ssl = SSL_new(ctx->ssl_ctx);
    if (session->ssl == NULL) {
        strcpy(error, "unable to create the required SSL object for the new connection");
        free_dtls_session(session);
        return NULL;
    }

    SSL_set_bio(session->ssl, bio, bio);
    SSL_set_accept_state(session->ssl);
    SSL_set_options(session->ssl, SSL_OP_COOKIE_EXCHANGE);

    if (!DTLS_set_link_mtu(session->ssl, DTLS_MAX_MTU)) {
        strcpy(error, "unable to set the MTU on the SSL object for the new connection");
        free_dtls_session(session);
        return NULL;
    }

    BIO_ADDR* peer_addr = BIO_ADDR_new();
    if (peer_addr == NULL) {
        strcpy(error, "unable to create the peer address");
        free_dtls_session(session);
        return NULL;
    }

    struct timeval timeout;
    timeout.tv_sec = 5;
    timeout.tv_usec = 0;
    BIO_ctrl(SSL_get_rbio(session->ssl), BIO_CTRL_DGRAM_SET_RECV_TIMEOUT, 0, &timeout);
    BIO_ctrl(SSL_get_rbio(session->ssl), BIO_CTRL_DGRAM_SET_SEND_TIMEOUT, 0, &timeout);

    int ret = 0;
    do { ret = DTLSv1_listen(session->ssl, peer_addr); }
    while (ret == 0);

    if (ret < 0) {
        strcpy(error, "unable to successfully preform CLIENT_HELLO/HELLO_VERIFY with the remote peer");
        BIO_ADDR_free(peer_addr);
        free_dtls_session(session);
        return NULL;
    }

    ret = 0;
    do { ret = SSL_accept(session->ssl); }
    while (ret == 0);

    if (ret < 0) {
        strcpy(error, "unable to successfully preform SSL_accept with the remote peer");
        BIO_ADDR_free(peer_addr);
        free_dtls_session(session);
        return NULL;
    }

    if (SSL_get_verify_result(session->ssl) != X509_V_OK) {
        strcpy(error, "unable to successfully verify the remote peer certificate");
        BIO_ADDR_free(peer_addr);
        free_dtls_session(session);
        return NULL;
    }

    if (ctx->use_v6) {
        session->remote_addr_len = sizeof(struct sockaddr_in6);
    } else {
        session->remote_addr_len = sizeof(struct sockaddr_in);
    }

    session->remote_addr = (struct sockaddr*)malloc(session->remote_addr_len);
    _parse_sock_addr(ctx->use_v6, BIO_ADDR_hostname_string(peer_addr, 1), htons(BIO_ADDR_rawport(peer_addr)), session->remote_addr);

    // Create the new connection socket.
    session->fd = socket(session->remote_addr->sa_family, SOCK_DGRAM, 0);

    // Set socket options on the new connection socket.
    const int on = 1, off = 0;
    setsockopt(session->fd, SOL_SOCKET, SO_REUSEADDR, (const void*) &on, (socklen_t) sizeof(on));
    if (ctx->use_v6) {
        setsockopt(session->fd, IPPROTO_IPV6, IPV6_V6ONLY, (char *)&off, sizeof(off));
    }

    // Bind the connection socket locally and connect back to the remote peer.
    if (bind(session->fd, ctx->local_addr, ctx->local_addr_len) < 0) {
        strcpy(error, "unable to successfully preform bind on the peer socket");
        BIO_ADDR_free(peer_addr);
        free_dtls_session(session);
        return NULL;
    }
    if (connect(session->fd, session->remote_addr, session->remote_addr_len) <0) {
        strcpy(error, "unable to successfully connect to the peer");
        BIO_ADDR_free(peer_addr);
        free_dtls_session(session);
        return NULL;
    }

    // Replace the BIO fd and set the BIO to connected.
    BIO_set_fd(SSL_get_rbio(session->ssl), session->fd, BIO_NOCLOSE);
    BIO_ctrl_set_connected(SSL_get_rbio(session->ssl), peer_addr);

    BIO_ADDR_free(peer_addr);

    return session;
}

Session* connect_dtls(Context* ctx, const char* addr, int port, char* error) {
    Session* session = (Session*)malloc(sizeof(Session));

    if (ctx->use_v6) {
        session->remote_addr_len = sizeof(struct sockaddr_in6);
    } else {
        session->remote_addr_len = sizeof(struct sockaddr_in);
    }

    session->remote_addr = (struct sockaddr*)malloc(session->remote_addr_len);
    _parse_sock_addr(ctx->use_v6, addr, port, session->remote_addr);

    session->fd = socket(ctx->local_addr->sa_family, SOCK_DGRAM, 0);

    const int on = 1, off = 0;
    setsockopt(session->fd, SOL_SOCKET, SO_REUSEADDR, (const void*) &on, (socklen_t) sizeof(on));
    if (ctx->use_v6) {
        setsockopt(session->fd, IPPROTO_IPV6, IPV6_V6ONLY, (char *)&off, sizeof(off));
    }

    // Bind the connection socket locally and connect back to the remote peer.
    if (bind(session->fd, ctx->local_addr, ctx->local_addr_len) < 0) {
        strcpy(error, "unable to successfully preform bind on the peer socket");
        free_dtls_session(session);
        return NULL;
    }
    if (connect(session->fd, session->remote_addr, session->remote_addr_len) <0) {
        strcpy(error, "unable to successfully connect to the peer");
        free_dtls_session(session);
        return NULL;
    }

    BIO* bio = BIO_new_dgram(session->fd, BIO_NOCLOSE);

    session->ssl = SSL_new(ctx->ssl_ctx);
    SSL_set_bio(session->ssl, bio, bio);
    SSL_set_connect_state(session->ssl);

    union BIO_sock_info_u peer_info;
    if ((peer_info.addr = BIO_ADDR_new()) == NULL) {
        strcpy(error, "unable to set the MTU on the SSL object for the new connection");
        free_dtls_session(session);
        return NULL;
    }
    if (!BIO_sock_info(session->fd, BIO_SOCK_INFO_ADDRESS, &peer_info)) {
        strcpy(error, "unable to set the MTU on the SSL object for the new connection");
        free_dtls_session(session);
        BIO_ADDR_free(peer_info.addr);
        return NULL;
    }

    BIO_ctrl_set_connected(bio, peer_info.addr);
    BIO_ADDR_free(peer_info.addr);
    peer_info.addr = NULL;

    if (!DTLS_set_link_mtu(session->ssl, DTLS_MAX_MTU)) {
        strcpy(error, "unable to set the MTU on the SSL object for the new connection");
        free_dtls_session(session);
        return NULL;
    }

    struct timeval timeout;
    timeout.tv_sec = 5;
    timeout.tv_usec = 0;
    BIO_ctrl(bio, BIO_CTRL_DGRAM_SET_RECV_TIMEOUT, 0, &timeout);
    BIO_ctrl(bio, BIO_CTRL_DGRAM_SET_SEND_TIMEOUT, 0, &timeout);

    if (!SSL_connect(session->ssl)) {
        strcpy(error, "failed to successfully preform SSL_connect with the remote peer");
        free_dtls_session(session);
        return NULL;
    }

    if (SSL_get_verify_result(session->ssl) != X509_V_OK) {
        strcpy(error, "unable to successfully verify the remote peer certificate");
        free_dtls_session(session);
        return NULL;
    }

    return session;
}

void free_dtls_session(Session* session) {
    if (session == NULL) {
        return;
    }

    if (session->ssl != NULL) {
        SSL_free(session->ssl);
        session->ssl = NULL;
    }

    if (session->fd >= 0) {
        close(session->fd);
        session->fd = -1;
    }
}

int get_dtls_fd(Session* session) {
    return session->fd;
}

int read_dtls(Session* session, void* buf, int length) {
    return SSL_read(session->ssl, buf, length);
}

int write_dtls(Session* session, void* buf, int length) {
    return SSL_write(session->ssl, buf, length);
}
