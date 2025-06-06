/* Copyright (c) 2010-2013 Marcus Westin */
'use strict';
(function (e, t) {
  typeof define == 'function' && define.amd
    ? define([], t)
    : typeof exports == 'object'
      ? (module.exports = t())
      : (e.store = t());
})(this, function () {
  function o() {
    try {
      return r in t && t[r];
    } catch (e) {
      return !1;
    }
  }
  var e = {},
    t = typeof window != 'undefined' ? window : global,
    n = t.document,
    r = 'localStorage',
    i = 'script',
    s;
  (e.disabled = !1),
    (e.version = '1.3.20'),
    (e.set = function (e, t) {}),
    (e.get = function (e, t) {}),
    (e.has = function (t) {
      return e.get(t) !== undefined;
    }),
    (e.remove = function (e) {}),
    (e.clear = function () {}),
    (e.transact = function (t, n, r) {
      r == null && ((r = n), (n = null)), n == null && (n = {});
      var i = e.get(t, n);
      r(i), e.set(t, i);
    }),
    (e.getAll = function () {}),
    (e.forEach = function () {}),
    (e.serialize = function (e) {
      return JSON.stringify(e);
    }),
    (e.deserialize = function (e) {
      if (typeof e != 'string') return undefined;
      try {
        return JSON.parse(e);
      } catch (t) {
        return e || undefined;
      }
    });
  if (o())
    (s = t[r]),
      (e.set = function (t, n) {
        return n === undefined ? e.remove(t) : (s.setItem(t, e.serialize(n)), n);
      }),
      (e.get = function (t, n) {
        var r = e.deserialize(s.getItem(t));
        return r === undefined ? n : r;
      }),
      (e.remove = function (e) {
        s.removeItem(e);
      }),
      (e.clear = function () {
        s.clear();
      }),
      (e.getAll = function () {
        var t = {};
        return (
          e.forEach(function (e, n) {
            t[e] = n;
          }),
          t
        );
      }),
      (e.forEach = function (t) {
        for (var n = 0; n < s.length; n++) {
          var r = s.key(n);
          t(r, e.get(r));
        }
      });
  else if (n && n.documentElement.addBehavior) {
    var u, a;
    try {
      (a = new ActiveXObject('htmlfile')),
        a.open(),
        a.write('<' + i + '>document.w=window</' + i + '><iframe src="/favicon.ico"></iframe>'),
        a.close(),
        (u = a.w.frames[0].document),
        (s = u.createElement('div'));
    } catch (f) {
      (s = n.createElement('div')), (u = n.body);
    }
    var l = function (t) {
        return function () {
          var n = Array.prototype.slice.call(arguments, 0);
          n.unshift(s), u.appendChild(s), s.addBehavior('#default#userData'), s.load(r);
          var i = t.apply(e, n);
          return u.removeChild(s), i;
        };
      },
      c = new RegExp('[!"#$%&\'()*+,/\\\\:;<=>?@[\\]^`{|}~]', 'g'),
      h = function (e) {
        return e.replace(/^d/, '___$&').replace(c, '___');
      };
    (e.set = l(function (t, n, i) {
      return (
        (n = h(n)),
        i === undefined ? e.remove(n) : (t.setAttribute(n, e.serialize(i)), t.save(r), i)
      );
    })),
      (e.get = l(function (t, n, r) {
        n = h(n);
        var i = e.deserialize(t.getAttribute(n));
        return i === undefined ? r : i;
      })),
      (e.remove = l(function (e, t) {
        (t = h(t)), e.removeAttribute(t), e.save(r);
      })),
      (e.clear = l(function (e) {
        var t = e.XMLDocument.documentElement.attributes;
        e.load(r);
        for (var n = t.length - 1; n >= 0; n--) e.removeAttribute(t[n].name);
        e.save(r);
      })),
      (e.getAll = function (t) {
        var n = {};
        return (
          e.forEach(function (e, t) {
            n[e] = t;
          }),
          n
        );
      }),
      (e.forEach = l(function (t, n) {
        var r = t.XMLDocument.documentElement.attributes;
        for (var i = 0, s; (s = r[i]); ++i) n(s.name, e.deserialize(t.getAttribute(s.name)));
      }));
  }
  try {
    var p = '__storejs__';
    e.set(p, p), e.get(p) != p && (e.disabled = !0), e.remove(p);
  } catch (f) {
    e.disabled = !0;
  }
  return (e.enabled = !e.disabled), e;
});
