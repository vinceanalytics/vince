!function(){"use strict";var e,i=window.location,o=window.document,l=o.getElementById("plausible"),s=l.getAttribute("data-api")||(e=(e=l).src.split("/"),c=e[0],e=e[2],c+"//"+e+"/api/event");function t(e,t){try{if("true"===window.localStorage.plausible_ignore)return a=t,(n="localStorage flag")&&console.warn("Ignoring Event: "+n),void(a&&a.callback&&a.callback())}catch(e){}var a,n={},r=(n.n=e,n.u=i.href,n.d=l.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),new XMLHttpRequest);r.open("POST",s,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=t;for(var n,r=0;r<a.length;r++)t.apply(this,a[r]);function p(){n!==i.pathname&&(n=i.pathname,t("pageview"))}var u,c=window.history;function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}c.pushState&&(u=c.pushState,c.pushState=function(){u.apply(this,arguments),p()},window.addEventListener("popstate",p)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){n||"visible"!==o.visibilityState||p()}):p();var d=1;function v(e){"auxclick"===e.type&&e.button!==d||((e=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target))&&e.href&&e.href.split("?")[0],function e(t,a){if(!t||b<a)return!1;if(w(t))return!0;return e(t.parentNode,a+1)}(e,0))}function m(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}function g(e){var e=w(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i)}return t}o.addEventListener("click",v),o.addEventListener("auxclick",v);var b=3;function h(e){if("auxclick"!==e.type||e.button===d){for(var t,a,n,r,i=e.target,o=0;o<=b&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;f(i)&&(t=i),w(i)&&(a=i),i=i.parentNode}a&&(r=g(a),t?(r.props.url=t.href,m(e,t,r)):((e={}).props=r.props,plausible(r.name,e)))}}function w(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,a=e.target,n=g(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),e={props:n.props,callback:r},plausible(n.name,e))}),o.addEventListener("click",h),o.addEventListener("auxclick",h)}();