#ifndef _tsdb_multi_language_support_h_
#define _tsdb_multi_language_support_h_

#include "stdint.h"

//<private>
/// version of current interface, even change compability, we must increment this version
//</private>
#define TSDB_ML_VERSION ((uint64_t)202120031650)
//<private>
/// version of lowest interface, for compability
//</private>
#define TSDB_ML_VERSION_LOW ((uint64_t)20210031650)

#ifdef __cplusplus
extern "C"
{
#endif

typedef unsigned char byte_t;
typedef int BOOL;

// Constants
#define TSDB_ML_RET_OK 0

// Forward declare
struct tsdb_v3_reader_t;
struct tsdb_v3_field_t;
struct tsdb_v3_t;
struct tsdb_ml_field_t;
struct tsdb_rows_t;
struct tsdb_result_set_t;
struct tsdb_ml_t;

typedef struct tsdb_v3_reader_t tsdb_v3_reader_t;
typedef struct tsdb_v3_field_t tsdb_v3_field_t;
typedef struct tsdb_ml_t tsdb_ml_t;
typedef void **tsdb_row_t;
typedef struct tsdb_ml_field_t tsdb_ml_field_t;
typedef struct tsdb_rows_t tsdb_rows_t;
typedef struct tsdb_result_set_t tsdb_result_set_t;
typedef tsdb_result_set_t RTDB_RES_SET;

struct tsdb_ml_field_t
{
    const char *name;
    uint16_t field_index;
    byte_t data_type;
    byte_t unique;
    byte_t has_index;
    byte_t is_ref;
    byte_t is_null;
    byte_t length;
    byte_t field_id;
    byte_t real_length;
    char _reserved[2];
};

struct tsdb_rows_t
{
    tsdb_rows_t *next;
    tsdb_row_t row;
    uint64_t len;
};

struct tsdb_result_set_t
{
    uint64_t row_count;
    uint32_t field_count;
    tsdb_v3_field_t **fields;
    tsdb_rows_t *data;
};

struct tsdb_ml_t
{
    uint64_t version;

    const char *build_version;

    void *inner_handle;

    void (*kill_me)(tsdb_ml_t *self);

    int (*connect)(const char *conn_str);

    int (*disconnect)();

    BOOL (*is_logined)
    ();

    const char *(*charset_get)();
    int (*charset_set)(const char *charset);

    const char *(*user_name)(tsdb_ml_t *self);

    const char *(*server_addr_str)(tsdb_ml_t *self);

    int (*pg_init)(const char *libpq_path, int *version);

    tsdb_v3_reader_t *(*table_new)(const char *type);

    int (*load_csv_file)(const char *path, tsdb_v3_reader_t *reader);

    const char *(*db_current)(tsdb_ml_t *self);

    int (*query)(tsdb_ml_t *self, const char *sql, int sql_len, const char *charset, const char *database);

    tsdb_v3_reader_t *(*store_result)(tsdb_ml_t *self);

    int (*test)(tsdb_ml_t *self, int argc, char **argv);

    //<private>
    /**
     * @brief
     * @param tsdb_ml_t *self
     * @param int req_bytes 请求包的字节大小
     * @param int rsp_bytes 响应包的字节大小
     * @warning 这个接口之前的函数声明是int (*call_test)(tsdb_ml_t *self, int bytes);为了兼容现在接口的改动做了修改
     */
    //</private>
    int (*call_test)(tsdb_ml_t *self, int req_bytes, int rsp_bytes);

    uint64_t (*affected_rows)(tsdb_ml_t *self);

    tsdb_v3_field_t **(*fetch_fields)(tsdb_ml_t *self);

    tsdb_ml_field_t **(*fetch_ml_fields)(tsdb_ml_t *self, int *field_count);

    tsdb_result_set_t *(*store_result_v2)(tsdb_ml_t *self);

    int (*free_result)(tsdb_ml_t *self, tsdb_result_set_t *result);
};

tsdb_ml_t *tsdb_ml_new_s(uint64_t version);

tsdb_ml_t *tsdb_ml_tls_s(uint64_t version);

/* Utility functions */
void destroy_tsdb_ml_fields(tsdb_ml_field_t **fields, uint32_t field_count);
tsdb_result_set_t *tsdb_result_set_constructor();

/* Functions interface */
tsdb_ml_t *tsdb_new();
tsdb_ml_t *tsdb_tls();
void tsdb_kill_me(void *self);
int tsdb_connect(const char *conn_str);
int tsdb_disconnect();
BOOL tsdb_is_logined();
const char *tsdb_charset_get();
int tsdb_charset_set(const char *charset);
const char *tsdb_user_name(void *self);
const char *tsdb_server_addr_str(void *self);
int tsdb_pg_init(const char *libpq_path, int *version);
tsdb_v3_reader_t *tsdb_table_new(const char *type);
int tsdb_load_csv_file(const char *path, tsdb_v3_reader_t *reader);
const char *tsdb_db_current(void *self);
int tsdb_query(void *self, const char *sql, int sql_len, const char *charset, const char *database);
uint64_t tsdb_affected_rows(void *self);
tsdb_v3_field_t **tsdb_fetch_fields(void *self);
tsdb_ml_field_t **tsdb_fetch_ml_fields(void *self, int *field_count);
tsdb_v3_reader_t *tsdb_store_result(void *self);
RTDB_RES_SET *tsdb_store_result_v2(void *self);
int tsdb_free_result(void *self, void *result);

#ifdef __cplusplus
}
#endif

#endif