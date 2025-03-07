/*
 * Copyright (c) 2013-2014 Conformal Systems <info@conformal.com>
 *
 * This file originated from: http://opensource.conformal.com/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

#ifndef __GLIB_GO_H__
#define __GLIB_GO_H__

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include <gio/gio.h>
#define G_SETTINGS_ENABLE_BACKEND
#include <gio/gsettingsbackend.h>
#include <glib-object.h>
#include <glib.h>
#include <glib/gi18n.h>
#include <locale.h>

/* GObject Type Casting */
static GType  objectGType   (GObject *obj) { return G_OBJECT_TYPE(obj); };

static GObject *toGObject(void *p) { return (G_OBJECT(p)); }

static GAction *toGAction(void *p) { return (G_ACTION(p)); }

static GActionGroup *toGActionGroup(void *p) { return (G_ACTION_GROUP(p)); }

static GActionMap *toGActionMap(void *p) { return (G_ACTION_MAP(p)); }

static GAsyncResult *toGAsyncResult(void *p) { return (G_ASYNC_RESULT(p)); }

static GSimpleAction *toGSimpleAction(void *p) { return (G_SIMPLE_ACTION(p)); }

static GSimpleActionGroup *toGSimpleActionGroup(void *p) {
  return (G_SIMPLE_ACTION_GROUP(p));
}

static GPropertyAction *toGPropertyAction(void *p) {
  return (G_PROPERTY_ACTION(p));
}

static GMenuModel *toGMenuModel(void *p) { return (G_MENU_MODEL(p)); }

static GMenu *toGMenu(void *p) { return (G_MENU(p)); }

static GMenuItem *toGMenuItem(void *p) { return (G_MENU_ITEM(p)); }

static GNotification *toGNotification(void *p) { return (G_NOTIFICATION(p)); }

static GPermission *toGPermission(void *p) { return (G_PERMISSION(p)); }

static GCancellable *toCancellable(void *p) { return (G_CANCELLABLE(p)); }

static GIcon *toGIcon(void *p) { return (G_ICON(p)); }
static GFileIcon *toGFileIcon(void *p) { return (G_FILE_ICON(p)); }

static GFile *toGFile(void *p) { return (G_FILE(p)); }

static GApplication *toGApplication(void *p) { return (G_APPLICATION(p)); }

static GSettings *toGSettings(void *p) { return (G_SETTINGS(p)); }

static GSettingsBackend *toGSettingsBackend(void *p) {
  return (G_SETTINGS_BACKEND(p));
}

static GBinding *toGBinding(void *p) { return (G_BINDING(p)); }

static GType _g_type_from_instance(gpointer instance) {
  return (G_TYPE_FROM_INSTANCE(instance));
}

/* Wrapper to avoid variable arg list */
static void _g_object_set_one(gpointer object, const gchar *property_name,
                              void *val) {
  g_object_set(object, property_name, *(gpointer **)val, NULL);
}

static GValue *_alloc_gvalue_list(int n) {
  GValue *valv;

  valv = g_new0(GValue, n);
  return (valv);
}

static void _val_list_insert(GValue *valv, int i, GValue *val) {
  valv[i] = *val;
}

/*
 * GValue
 */

static GValue *_g_value_alloc() { return (g_new0(GValue, 1)); }

static GValue *_g_value_init(GType g_type) {
  GValue *value;

  value = g_new0(GValue, 1);
  return (g_value_init(value, g_type));
}

static gboolean _g_type_is_value(GType g_type) {
  return (G_TYPE_IS_VALUE(g_type));
}

static gboolean _g_is_value(GValue *val) { return (G_IS_VALUE(val)); }

static GType _g_value_type(GValue *val) { return (G_VALUE_TYPE(val)); }

static const gchar *_g_value_type_name(GValue *val) {
  return (G_VALUE_TYPE_NAME(val));
}

static GType _g_value_fundamental(GType type) {
  return (G_TYPE_FUNDAMENTAL(type));
}

static GObjectClass *_g_object_get_class(GObject *object) {
  return (G_OBJECT_GET_CLASS(object));
}

/*
 * Closure support
 */

extern void goMarshal(GClosure *, GValue *, guint, GValue *, gpointer, gpointer);

extern void removeClosure(gpointer, GClosure *);

/**
 * create a new closure, handle must be a stored gopointer to a *closureContext
 */
static GClosure *_g_closure_new(gpointer handle)
{
  GClosure *closure;

  closure = g_closure_new_simple(sizeof(GClosure), NULL);
  g_closure_set_meta_marshal(closure, handle, (GClosureMarshal)(goMarshal));
  g_closure_add_finalize_notifier(closure, handle, removeClosure);

  return (closure);
}

static inline guint _g_signal_new(const gchar *name) {
  return g_signal_new(name, G_TYPE_OBJECT, G_SIGNAL_RUN_FIRST | G_SIGNAL_ACTION,
                      0, NULL, NULL, g_cclosure_marshal_VOID__POINTER,
                      G_TYPE_NONE, 0);
}

static void init_i18n(const char *domain, const char *dir) {
  setlocale(LC_ALL, "");
  bindtextdomain(domain, dir);
  bind_textdomain_codeset(domain, "UTF-8");
  textdomain(domain);
}

static const char *localize(const char *string) { return _(string); }

static inline char **make_strings(int count) {
  return (char **)malloc(sizeof(char *) * count);
}

static inline void destroy_strings(char **strings) { free(strings); }

static inline char *get_string(char **strings, int n) { return strings[n]; }

static inline void set_string(char **strings, int n, char *str) {
  strings[n] = str;
}

static inline gchar **next_gcharptr(gchar **s) { return (s + 1); }

extern void goCompareDataFuncs(gconstpointer a, gconstpointer b,
                               gpointer user_data);


/**
 * custom glib type for arbitrary go data
 */

#define GLIB_GO_TYPE_ARBITRARY_DATA (glib_go_arbitrary_data_get_type())

typedef struct
{
  gpointer data; // Arbitrary data, corresponds to a gopointer handle in Go
} GlibGoArbitraryData;

GType glib_go_arbitrary_data_get_type(void);

static GlibGoArbitraryData *glib_go_arbitrary_data_new(gpointer data);
static GlibGoArbitraryData *glib_go_arbitrary_data_copy (GlibGoArbitraryData * orig);
static void glib_go_arbitrary_data_free (GlibGoArbitraryData * orig);

/**
 * end custom glib type for arbitrary go data
 */

#endif
