!function(){"use strict";var t,e,n,o=window.location,c=window.document,s=c.getElementById("vince"),l=s.getAttribute("data-api")||(t=s.src.split("/"),e=t[0],n=t[2],e+"//"+n+"/api/event");function d(t){console.warn("Ignoring Event: "+t)}function i(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return d("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return d("localStorage flag")}catch(t){}var n={};n.n=t,n.u=o.href,n.d=s.getAttribute("data-domain"),n.r=c.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props);var i=s.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)}),a=n.p||{};i.forEach(function(t){var e=t.replace("event-",""),n=s.getAttribute(t);a[e]=a[e]||n}),n.p=a,n.h=1;var r=new XMLHttpRequest;r.open("POST",l,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&e&&e.callback&&e.callback()}}}var a=window.vince&&window.vince.q||[];window.vince=i;for(var r,w=0;w<a.length;w++)i.apply(this,a[w]);function v(){r=o.pathname,i("pageview")}window.addEventListener("hashchange",v),"prerender"===c.visibilityState?c.addEventListener("visibilitychange",function(){r||"visible"!==c.visibilityState||v()}):v()}();