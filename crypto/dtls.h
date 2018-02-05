#ifndef dtls_h_loaded
#define dtls_h_loaded

// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

#include <errno.h>
#include <unistd.h>
#include <netinet/in.h>
#include <string.h>
#include <arpa/inet.h>
#include <openssl/ssl.h>
#include <openssl/conf.h>
#include <openssl/engine.h>
#include <sys/socket.h>
#include <netdb.h>

#define COOKIE_SECRET_LENGTH 16
#define SSL_CIPHER "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384"
#define CLIENT_CTX_PORT 0
#define DTLS_MAX_MTU 1000

typedef struct {
	int fd;
    SSL_CTX* ssl_ctx;
    int use_v6;

    struct sockaddr* local_addr;
    unsigned int local_addr_len;
} Context;

typedef struct{
	int fd;
	SSL* ssl;

	struct sockaddr* remote_addr;
	unsigned int remote_addr_len;
} Session;

void init_dtls();
void destroy_dtls();

Context* init_server_dtls_context(int fd, const char* addr, int port, int use_v6, int verify_peer, const char* ca, const char* cert, const char* key, char* error);
Context* init_client_dtls_context(const char* addr, int use_v6, int verify_peer, const char* ca, const char* cert, const char* key, char* error);
void free_dtls_context(Context* ctx);

Session* accept_dtls(Context* ctx, char* error);
Session* connect_dtls(Context* ctx, const char* addr, int port, char* error);

int get_dtls_fd(Session* session);

int read_dtls(Session* session, void* buf, int length);
int write_dtls(Session* session, void* buf, int length);
void free_dtls_session(Session* session);
#endif
