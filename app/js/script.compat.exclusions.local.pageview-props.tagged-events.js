!function(){"use strict";var e,l=window.location,u=window.document,s=u.getElementById("plausible"),p=s.getAttribute("data-api")||(e=(e=s).src.split("/"),f=e[0],e=e[2],f+"//"+e+"/api/event");function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function t(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var a=s&&s.getAttribute("data-include"),n=s&&s.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(r),n=n&&n.split(",").some(r);if(!a||n)return c("exclusion rule",t)}function r(e){return l.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},n=(a.n=e,a.u=l.href,a.d=s.getAttribute("data-domain"),a.r=u.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=a.p||{},o=(n.forEach(function(e){var t=e.replace("event-",""),e=s.getAttribute(e);i[t]=i[t]||e}),a.p=i,new XMLHttpRequest);o.open("POST",p,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(a)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=t;for(var n,r=0;r<a.length;r++)t.apply(this,a[r]);function i(){n!==l.pathname&&(n=l.pathname,t("pageview"))}var o,f=window.history;function d(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}f.pushState&&(o=f.pushState,f.pushState=function(){o.apply(this,arguments),i()},window.addEventListener("popstate",i)),"prerender"===u.visibilityState?u.addEventListener("visibilitychange",function(){n||"visible"!==u.visibilityState||i()}):i();var v=1;function m(e){"auxclick"===e.type&&e.button!==v||((e=function(e){for(;e&&(void 0===e.tagName||!d(e)||!e.href);)e=e.parentNode;return e}(e.target))&&e.href&&e.href.split("?")[0],function e(t,a){if(!t||h<a)return!1;if(y(t))return!0;return e(t.parentNode,a+1)}(e,0))}function g(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}function b(e){var e=y(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i)}return t}u.addEventListener("click",m),u.addEventListener("auxclick",m);var h=3;function w(e){if("auxclick"!==e.type||e.button===v){for(var t,a,n,r,i=e.target,o=0;o<=h&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;d(i)&&(t=i),y(i)&&(a=i),i=i.parentNode}a&&(r=b(a),t?(r.props.url=t.href,g(e,t,r)):((e={}).props=r.props,plausible(r.name,e)))}}function y(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}u.addEventListener("submit",function(e){var t,a=e.target,n=b(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),e={props:n.props,callback:r},plausible(n.name,e))}),u.addEventListener("click",w),u.addEventListener("auxclick",w)}();